package events_test

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/ledger"
)

// queue declares a consumer queue of the test's own, bound under its own name, and deletes it and its
// dead-letter queue when the test ends.
func queue(t *testing.T, conn *amqp.Connection) string {
	t.Helper()
	name := "test.consume." + strings.ToLower(rand.Text())
	ch := channel(t, conn, false)
	if err := events.DeclareQueue(ch, name, name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, q := range []string{name, name + ".dead"} {
			if _, err := ch.QueueDelete(q, false, false, false); err != nil {
				t.Logf("cleanup %s: %v", q, err)
			}
		}
	})
	return name
}

// calls is a handler that records each message id it is run for, failing the ones in fail.
type calls struct {
	mu   sync.Mutex
	runs map[string]int
	fail map[string]bool
}

func (c *calls) handle(_ context.Context, _ pgx.Tx, d amqp.Delivery) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.runs[d.MessageId]++
	if c.fail[d.MessageId] {
		return errors.New("poison")
	}
	return nil
}

func (c *calls) count(id string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.runs[id]
}

// consume runs events.Consume on queue with handle and returns a stop that waits until Consume has returned.
func consume(t *testing.T, db *pgxpool.Pool, url, queue string, handle events.Handler) func() {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		events.Consume(ctx, db, url, queue, []string{queue}, handle)
		close(done)
	}()
	stop := func() {
		cancel()
		<-done
	}
	t.Cleanup(stop)
	return stop
}

func publish(t *testing.T, conn *amqp.Connection, key string, ids ...string) {
	t.Helper()
	ch := channel(t, conn, false)
	for _, id := range ids {
		if err := ch.PublishWithContext(context.Background(), events.Exchange, key, false, false, amqp.Publishing{MessageId: id, Body: []byte(id)}); err != nil {
			t.Fatal(err)
		}
	}
}

// eventually polls ok every 100 ms for up to 15 s.
func eventually(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(15 * time.Second); !ok(); time.Sleep(100 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("after 15 s: %s", what)
		}
	}
}

func waiting(t *testing.T, conn *amqp.Connection, queue string) int {
	t.Helper()
	ch := channel(t, conn, false)
	defer ch.Close()
	q, err := ch.QueueDeclarePassive(queue, true, false, false, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	return q.Messages
}

// settled is true once id is committed in processed_events and the pool holds no acquired connection: Process has
// left its transaction without error, so its ack goes out before Consume returns.
func settled(t *testing.T, db *pgxpool.Pool, id string) bool {
	t.Helper()
	var n int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM processed_events WHERE message_id = $1`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n == 1 && db.Stat().AcquiredConns() == 0
}

// A message delivered twice runs its handler once, and both deliveries are acked (§4.3, spec #92): the
// message behind them, processed in order, proves the second was handled before the consumer stops.
func TestADuplicateDeliveryIsProcessedOnce(t *testing.T) {
	db := scratch(t, "LEDGER_TEST_DATABASE_URL", ledger.Migrate)
	url, conn := broker(t)
	q := queue(t, conn)
	c := &calls{runs: map[string]int{}}
	stop := consume(t, db, url, q, c.handle)
	publish(t, conn, q, "twice", "twice", "after")
	eventually(t, "the message behind the duplicate is committed and its ack sent", func() bool { return settled(t, db, "after") })
	stop()
	if c.count("twice") != 1 {
		t.Errorf("the handler ran %d times for a message delivered twice, want 1", c.count("twice"))
	}
	var seen int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM processed_events WHERE message_id IN ('twice', 'after')`).Scan(&seen); err != nil || seen != 2 {
		t.Errorf("processed_events holds %d of the two ids (%v), want both once", seen, err)
	}
	if n := waiting(t, conn, q); n != 0 {
		t.Errorf("%d messages back on the queue after the consumer stopped, want every delivery acked", n)
	}
}

// A message that keeps failing stops after five tries and waits in <queue>.dead, as does one with no message
// id, while the message behind them is processed (§4.3).
func TestAPoisonMessageWaitsForAPerson(t *testing.T) {
	db := scratch(t, "LEDGER_TEST_DATABASE_URL", ledger.Migrate)
	url, conn := broker(t)
	q := queue(t, conn)
	c := &calls{runs: map[string]int{}, fail: map[string]bool{"poison": true}}
	consume(t, db, url, q, c.handle)
	publish(t, conn, q, "poison", "", "good")
	eventually(t, "both bad messages reach the dead-letter queue", func() bool { return waiting(t, conn, q+".dead") == 2 })
	if c.count("poison") != 5 || c.count("good") != 1 || c.count("") != 0 {
		t.Errorf("the handler ran %d times for the poison, %d for the good one, %d for no id; want 5, 1 and 0", c.count("poison"), c.count("good"), c.count(""))
	}
	var processed int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM processed_events`).Scan(&processed); err != nil || processed != 1 {
		t.Errorf("processed_events holds %d ids (%v), want only the good one", processed, err)
	}
}

// ackOnly answers the consumer's ack or reject for one delivery handed to Process directly.
type ackOnly struct{}

func (ackOnly) Ack(uint64, bool) error        { return nil }
func (ackOnly) Nack(uint64, bool, bool) error { return errors.New("nacked") }
func (ackOnly) Reject(uint64, bool) error     { return nil }

// A function handed to AfterCommit runs once its handler's transaction commits and never after a rollback, and
// AfterCommit outside Process does nothing (§4.3: notify pushes only a committed stamp).
func TestAfterCommitRunsOnlyOnACommit(t *testing.T) {
	db := scratch(t, "LEDGER_TEST_DATABASE_URL", ledger.Migrate)
	ran := map[string]int{}
	handle := func(ctx context.Context, _ pgx.Tx, d amqp.Delivery) error {
		events.AfterCommit(ctx, func() { ran[d.MessageId]++ })
		if d.MessageId == "rolled-back" {
			return errors.New("poison")
		}
		return nil
	}
	for _, id := range []string{"committed", "rolled-back"} {
		events.Process(t.Context(), db, amqp.Delivery{Acknowledger: ackOnly{}, MessageId: id}, handle)
	}
	events.AfterCommit(t.Context(), func() { ran["outside"]++ })
	if ran["committed"] != 1 || ran["rolled-back"] != 0 || ran["outside"] != 0 {
		t.Errorf("ran %v, want only the committed delivery's function, once", ran)
	}
}
