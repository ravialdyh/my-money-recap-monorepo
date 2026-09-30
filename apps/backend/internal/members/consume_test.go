package members

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
)

// acks counts what the consumer answered the broker for one delivery; a nack fails, since only a reject counts
// toward the five tries.
type acks struct{ acked, rejected int }

func (a *acks) Ack(uint64, bool) error { a.acked++; return nil }

func (a *acks) Nack(uint64, bool, bool) error { return errors.New("the consumer nacked") }

func (a *acks) Reject(uint64, bool) error { a.rejected++; return nil }

// ledger.currency.converted sets the Currency once per message id and stages one members.updated; a conversion
// older than the one applied, or a malformed event, changes nothing, the first acked and the second rejected.
func TestTheConsumerTakesTheConvertedCurrency(t *testing.T) {
	db := testDB(t)
	subject := "consumer-" + rand.Text()
	var ids []string
	t.Cleanup(func() {
		ctx := context.Background()
		for _, q := range []string{`DELETE FROM outbox WHERE member_id IN (SELECT id FROM member WHERE keycloak_subject = $1)`, `DELETE FROM member WHERE keycloak_subject = $1`} {
			if _, err := db.Exec(ctx, q, subject); err != nil {
				t.Logf("cleanup: %v", err)
			}
		}
		if _, err := db.Exec(ctx, `DELETE FROM processed_events WHERE message_id = ANY($1)`, ids); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject})
	m, err := Upsert(ctx, db, subject, "c@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE member SET currency_code = 'GBP' WHERE id = $1`, m.ID); err != nil {
		t.Fatal(err)
	}
	base := time.Now()
	deliver := func(messageID string, msg proto.Message) bool {
		t.Helper()
		body, err := proto.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, messageID)
		a := &acks{}
		events.Process(t.Context(), db, amqp.Delivery{Acknowledger: a, MessageId: messageID, RoutingKey: Keys[0], Body: body}, ApplyConversion)
		return a.acked == 1 && a.rejected == 0
	}
	converted := func(id string, at time.Duration, to string) *moneyrecapv1.LedgerCurrencyConverted {
		return &moneyrecapv1.LedgerCurrencyConverted{
			Envelope: &moneyrecapv1.EventEnvelope{MemberId: id, OccurredAt: timestamppb.New(base.Add(at))}, ToCurrencyCode: to, Rate: "1", Source: moneyrecapv1.RateSource_RATE_SOURCE_MANUAL,
		}
	}
	currency := func() string {
		t.Helper()
		row, err := Get(ctx, db, m.ID)
		if err != nil {
			t.Fatal(err)
		}
		return row.CurrencyCode
	}
	updates := func() int {
		n := 0
		for _, s := range outbox(t, db, m.ID) {
			if s.key == "members.updated" {
				n++
			}
		}
		return n
	}

	once := rand.Text()
	if !deliver(once, converted(m.ID, time.Minute, "IDR")) || currency() != "IDR" || updates() != 1 {
		t.Fatalf("the conversion: Currency %s, %d members.updated, want IDR and one", currency(), updates())
	}
	if !deliver(once, converted(m.ID, 2*time.Minute, "USD")) || currency() != "IDR" || updates() != 1 {
		t.Fatalf("the same message again: Currency %s, %d members.updated, want nothing changed and an ack", currency(), updates())
	}
	if !deliver(rand.Text(), converted(m.ID, 0, "EUR")) || currency() != "IDR" || updates() != 1 {
		t.Fatalf("an older conversion arriving late: Currency %s, %d members.updated, want nothing changed and an ack", currency(), updates())
	}
	if !deliver(rand.Text(), converted("00000000-0000-4000-8000-00000000c173", time.Hour, "USD")) {
		t.Error("a conversion for a Member members no longer holds was not acked")
	}
	for name, msg := range map[string]proto.Message{
		"no Member":    converted("", time.Hour, "USD"),
		"no Currency":  converted(m.ID, time.Hour, "XXX"),
		"no time":      &moneyrecapv1.LedgerCurrencyConverted{Envelope: &moneyrecapv1.EventEnvelope{MemberId: m.ID}, ToCurrencyCode: "USD"},
		"not an event": &moneyrecapv1.MembersUpdated{Envelope: &moneyrecapv1.EventEnvelope{MemberId: m.ID}},
	} {
		if deliver(rand.Text(), msg) {
			t.Errorf("%s: acked, want a reject", name)
		}
	}
	if got := outbox(t, db, m.ID); got[len(got)-1].event.Facts.GetCurrencyCode() != "IDR" {
		t.Errorf("members.updated carries %v, want IDR", got[len(got)-1].event.Facts)
	}
}
