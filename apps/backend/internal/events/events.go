// Package events is ARCHITECTURE §4.3 for every service: the topology on RabbitMQ, the outbox row a change
// writes in its own transaction, the relay that publishes those rows, and the consumer that processes each
// delivery once in effect.
package events

import (
	"context"
	"errors"
	"log"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// The names and limits §4.3 gives: one topic exchange, the dead-letter exchange, five tries before a message
// waits for a person, and sixteen unacked deliveries at a time.
const (
	Exchange    = "mmr.events"
	DeadLetter  = "mmr.dlx"
	tries       = 5
	prefetch    = 16
	pollEvery   = time.Second
	redialAfter = 5 * time.Second
	batch       = 50
)

// Stage writes one outbox row inside the caller's transaction, so the change and its event commit or roll back
// together (A4). build receives the envelope, whose message id is the row's id.
func Stage(ctx context.Context, tx pgx.Tx, memberID, routingKey string, build func(*moneyrecapv1.EventEnvelope) proto.Message) error {
	env := &moneyrecapv1.EventEnvelope{MessageId: uuid.NewV4().String(), OccurredAt: timestamppb.Now(), MemberId: memberID, RoutingKey: routingKey}
	payload, err := proto.Marshal(build(env))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox (id, member_id, routing_key, payload) VALUES ($1, $2, $3, $4)`, env.MessageId, memberID, routingKey, payload)
	return err
}

// Declare makes the two exchanges, idempotently; every service declares what it leans on.
func Declare(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	return ch.ExchangeDeclare(DeadLetter, "topic", true, false, false, false, nil)
}

// DeclareQueue makes a consumer's quorum queue bound to keys, and its <queue>.dead queue, where a delivery lands
// after five tries. RabbitMQ's x-delivery-limit counts the failures it forgives, so five tries is a limit of 4.
func DeclareQueue(ch *amqp.Channel, queue string, keys ...string) error {
	if err := Declare(ch); err != nil {
		return err
	}
	dead := queue + ".dead"
	if _, err := ch.QueueDeclare(dead, true, false, false, false, amqp.Table{"x-queue-type": "quorum"}); err != nil {
		return err
	}
	if err := ch.QueueBind(dead, queue, DeadLetter, false, nil); err != nil {
		return err
	}
	args := amqp.Table{"x-queue-type": "quorum", "x-delivery-limit": tries - 1, "x-dead-letter-exchange": DeadLetter, "x-dead-letter-routing-key": queue}
	if _, err := ch.QueueDeclare(queue, true, false, false, false, args); err != nil {
		return err
	}
	for _, key := range keys {
		if err := ch.QueueBind(queue, key, Exchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}

// Relay publishes the service's unsent outbox rows every second and returns when ctx ends, dialling again after
// a lost connection.
func Relay(ctx context.Context, db *pgxpool.Pool, url string) {
	for ctx.Err() == nil {
		if err := relay(ctx, db, url); err != nil && ctx.Err() == nil {
			log.Printf("outbox relay: %v", err)
			wait(ctx, redialAfter)
		}
	}
}

func relay(ctx context.Context, db *pgxpool.Pool, url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := ch.Confirm(false); err != nil {
		return err
	}
	if err := Declare(ch); err != nil {
		return err
	}
	for ctx.Err() == nil {
		if err := PublishPending(ctx, db, ch); err != nil {
			return err
		}
		wait(ctx, pollEvery)
	}
	return nil
}

// PublishPending sends up to fifty unsent rows with publisher confirms and marks sent only the ones the broker
// confirmed (§4.3); a refused row waits for the next poll without holding up the rest. FOR UPDATE SKIP LOCKED
// keeps two relays off the same row.
func PublishPending(ctx context.Context, db *pgxpool.Pool, ch *amqp.Channel) error {
	return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT id::text, routing_key, payload FROM outbox WHERE sent_at IS NULL ORDER BY created_at, id LIMIT $1 FOR UPDATE SKIP LOCKED`, batch)
		if err != nil {
			return err
		}
		pending, err := pgx.CollectRows(rows, pgx.RowToStructByPos[outboxRow])
		if err != nil {
			return err
		}
		var sent []string
		for _, r := range pending {
			confirm, err := ch.PublishWithDeferredConfirmWithContext(ctx, Exchange, r.Key, false, false, amqp.Publishing{
				MessageId: r.ID, DeliveryMode: amqp.Persistent, ContentType: "application/x-protobuf", Body: r.Payload,
			})
			if err != nil {
				return err
			}
			acked, err := confirm.WaitContext(ctx)
			if err != nil {
				break
			}
			if acked {
				sent = append(sent, r.ID)
			}
		}
		_, err = tx.Exec(ctx, `UPDATE outbox SET sent_at = now() WHERE id = ANY($1::uuid[])`, sent)
		return err
	})
}

type outboxRow struct {
	ID, Key string
	Payload []byte
}

// Handler applies one delivery inside the transaction that records it as processed.
type Handler func(ctx context.Context, tx pgx.Tx, d amqp.Delivery) error

// committed carries the functions a Handler hands AfterCommit through its ctx.
type committed struct{ after []func() }

type committedKey struct{}

// AfterCommit runs fn once the transaction of the Handler whose ctx this is has committed, after the ack, and never
// when it rolls back. Outside Process it does nothing.
func AfterCommit(ctx context.Context, fn func()) {
	if c, ok := ctx.Value(committedKey{}).(*committed); ok {
		c.after = append(c.after, fn)
	}
}

// errNoMessageID refuses a delivery with no message id, which processed_events could not tell from another.
var errNoMessageID = errors.New("the delivery carries no message id")

// Consume processes queue's deliveries until ctx ends, dialling again after a lost connection.
func Consume(ctx context.Context, db *pgxpool.Pool, url, queue string, keys []string, handle Handler) {
	for ctx.Err() == nil {
		if err := consume(ctx, db, url, queue, keys, handle); err != nil && ctx.Err() == nil {
			log.Printf("consumer %s: %v", queue, err)
			wait(ctx, redialAfter)
		}
	}
}

func consume(ctx context.Context, db *pgxpool.Pool, url, queue string, keys []string, handle Handler) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := DeclareQueue(ch, queue, keys...); err != nil {
		return err
	}
	if err := ch.Qos(prefetch, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.ConsumeWithContext(ctx, queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for d := range deliveries {
		Process(ctx, db, d, handle)
	}
	return errors.New("the broker closed the deliveries")
}

// Process runs handle once in effect (§4.3): its transaction first records the message id in processed_events,
// so a repeat is only acked. A failure is rejected, never nacked: RabbitMQ 4.3 counts only a reject toward the
// five tries.
func Process(ctx context.Context, db *pgxpool.Pool, d amqp.Delivery, handle Handler) {
	c := &committed{}
	ctx = context.WithValue(ctx, committedKey{}, c)
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		if d.MessageId == "" {
			return errNoMessageID
		}
		seen, err := tx.Exec(ctx, `INSERT INTO processed_events (message_id) VALUES ($1) ON CONFLICT DO NOTHING`, d.MessageId)
		if err != nil || seen.RowsAffected() == 0 {
			return err
		}
		return handle(ctx, tx, d)
	})
	if err != nil {
		log.Printf("event %s on %s: %v", d.MessageId, d.RoutingKey, err)
		_ = d.Reject(true)
		return
	}
	_ = d.Ack(false)
	for _, fn := range c.after {
		fn()
	}
}

func wait(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
