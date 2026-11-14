package events_test

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/members"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testdb"
)

// scratch is testdb.Scratch: the stack's relay and consumer never meet a test's rows.
func scratch(t *testing.T, dsnVar string, migrate func(context.Context, *pgxpool.Pool) error) *pgxpool.Pool {
	t.Helper()
	return testdb.Scratch(t, dsnVar, migrate)
}

// broker dials the Compose stack's RabbitMQ and declares the exchanges, skipping unless MMR_TEST_AMQP_URL is set.
func broker(t *testing.T) (string, *amqp.Connection) {
	t.Helper()
	url := os.Getenv("MMR_TEST_AMQP_URL")
	if url == "" {
		t.Skip("MMR_TEST_AMQP_URL not set")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := events.Declare(channel(t, conn, false)); err != nil {
		t.Fatal(err)
	}
	return url, conn
}

func channel(t *testing.T, conn *amqp.Connection, confirm bool) *amqp.Channel {
	t.Helper()
	ch, err := conn.Channel()
	if err == nil && confirm {
		err = ch.Confirm(false)
	}
	if err != nil {
		t.Fatal(err)
	}
	return ch
}

func updated(env *moneyrecapv1.EventEnvelope) proto.Message {
	return &moneyrecapv1.MembersUpdated{Envelope: env, Facts: &moneyrecapv1.MemberFacts{Timezone: "Asia/Jakarta"}}
}

// stage commits n outbox rows for one Member under key, one transaction each, and returns their ids.
func stage(t *testing.T, db *pgxpool.Pool, key string, n int) []string {
	t.Helper()
	ctx, member := context.Background(), uuid.New().String()
	for range n {
		if err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error { return events.Stage(ctx, tx, member, key, updated) }); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := db.Query(ctx, `SELECT id::text FROM outbox WHERE routing_key = $1 ORDER BY created_at, id`, key)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	return ids
}

func unsent(t *testing.T, db *pgxpool.Pool) int {
	t.Helper()
	var n int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM outbox WHERE sent_at IS NULL`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// listen binds a queue of the test's own to key and returns a wait for n messages that then counts, per
// message id, everything that arrived, including any duplicate within the following second.
func listen(t *testing.T, conn *amqp.Connection, key string) func(n int) map[string]int {
	t.Helper()
	ch := channel(t, conn, false)
	q, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err == nil {
		err = ch.QueueBind(q.Name, key, events.Exchange, false, nil)
	}
	deliveries, cerr := ch.Consume(q.Name, "", true, true, false, false, nil)
	if err = errors.Join(err, cerr); err != nil {
		t.Fatal(err)
	}
	return func(n int) map[string]int {
		seen := map[string]int{}
		var grace <-chan time.Time
		for got := 0; ; {
			select {
			case d := <-deliveries:
				seen[d.MessageId]++
				got++
				if got == n {
					grace = time.After(time.Second)
				}
			case <-grace:
				return seen
			case <-time.After(15 * time.Second):
				t.Fatalf("%d of %d messages, then nothing for 15 s", got, n)
			}
		}
	}
}

// A change and its event commit together: a rolled-back write leaves no row, and a committed one leaves one
// whose payload names its own row as the message id (A4, §4.3).
func TestStageCommitsWithItsChange(t *testing.T) {
	db := scratch(t, "MMR_TEST_MEMBERS_DATABASE_URL", members.Migrate)
	ctx, member := context.Background(), uuid.New().String()
	failed := errors.New("the change failed")
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		return errors.Join(events.Stage(ctx, tx, member, "members.updated", updated), failed)
	})
	if !errors.Is(err, failed) || unsent(t, db) != 0 {
		t.Fatalf("a rolled-back change: %v, %d rows, want its error and none", err, unsent(t, db))
	}
	if err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error { return events.Stage(ctx, tx, member, "members.updated", updated) }); err != nil {
		t.Fatal(err)
	}
	var id, owner, key string
	var payload []byte
	if err := db.QueryRow(ctx, `SELECT id::text, member_id::text, routing_key, payload FROM outbox`).Scan(&id, &owner, &key, &payload); err != nil {
		t.Fatal(err)
	}
	var e moneyrecapv1.MembersUpdated
	if err := proto.Unmarshal(payload, &e); err != nil {
		t.Fatal(err)
	}
	env := e.GetEnvelope()
	if env.GetMessageId() != id || env.GetMemberId() != member || owner != member || env.GetRoutingKey() != key || env.GetOccurredAt() == nil || e.GetFacts().GetTimezone() != "Asia/Jakarta" {
		t.Fatalf("row %s for %s under %s carries %v", id, owner, key, &e)
	}
}

// Two relays on one table never publish a row twice (§4.3): a relay mid-batch holds its rows, which another
// skips rather than waits for or sends too, and two draining together deliver every row exactly once.
func TestRelaysPublishEachRowOnce(t *testing.T) {
	db := scratch(t, "MMR_TEST_MEMBERS_DATABASE_URL", members.Migrate)
	_, conn := broker(t)
	ctx, key := context.Background(), "test.relay."+strings.ToLower(rand.Text())
	received := listen(t, conn, key)
	ids := stage(t, db, key, 120)

	held, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := held.Exec(ctx, `SELECT id FROM outbox ORDER BY created_at, id LIMIT 10 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}
	if err := events.PublishPending(ctx, db, channel(t, conn, true)); err != nil {
		t.Fatal(err)
	}
	var heldUnsent int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE sent_at IS NULL AND id = ANY($1::uuid[])`, ids[:10]).Scan(&heldUnsent); err != nil {
		t.Fatal(err)
	}
	if n := unsent(t, db); n != 70 || heldUnsent != 10 {
		t.Fatalf("one batch beside a relay holding 10 rows left %d unsent, %d of them held; want 70 and 10", n, heldUnsent)
	}
	if err := held.Rollback(ctx); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 2 {
		ch := channel(t, conn, true)
		wg.Go(func() {
			for range 3 {
				if err := events.PublishPending(ctx, db, ch); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	if n := unsent(t, db); n != 0 {
		t.Fatalf("%d rows unsent after two relays drained the table", n)
	}
	seen := received(len(ids))
	for _, id := range ids {
		if seen[id] != 1 {
			t.Errorf("row %s published %d times", id, seen[id])
		}
	}
	if len(seen) != len(ids) {
		t.Errorf("%d distinct messages for %d rows", len(seen), len(ids))
	}
}

// A publish the broker refuses leaves its row unsent for the next poll, and the rows behind it still go:
// only a confirmed row is marked (§4.3). The refusal comes from a full queue that rejects new messages.
func TestARefusedPublishStaysUnsent(t *testing.T) {
	db := scratch(t, "MMR_TEST_MEMBERS_DATABASE_URL", members.Migrate)
	_, conn := broker(t)
	ch := channel(t, conn, false)
	refused, accepted := "test.refused."+strings.ToLower(rand.Text()), "test.accepted."+strings.ToLower(rand.Text())
	q, err := ch.QueueDeclare("", false, true, true, false, amqp.Table{"x-max-length": 0, "x-overflow": "reject-publish"})
	if err == nil {
		err = ch.QueueBind(q.Name, refused, events.Exchange, false, nil)
	}
	if err != nil {
		t.Fatal(err)
	}
	received := listen(t, conn, accepted)
	stage(t, db, refused, 1)
	behind := stage(t, db, accepted, 2)
	if err := events.PublishPending(context.Background(), db, channel(t, conn, true)); err != nil {
		t.Fatal(err)
	}
	if n := unsent(t, db); n != 1 {
		t.Fatalf("%d rows unsent after the broker refused one of three, want 1", n)
	}
	if seen := received(len(behind)); seen[behind[0]] != 1 || seen[behind[1]] != 1 {
		t.Fatalf("the two rows behind the refused one arrived %v, want once each", seen)
	}
}
