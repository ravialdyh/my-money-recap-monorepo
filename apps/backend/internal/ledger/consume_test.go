package ledger

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// apply runs ApplyMemberFacts on one delivery of msg in a transaction of its own, as events.Process does.
func apply(t *testing.T, db *pgxpool.Pool, msg proto.Message) error {
	t.Helper()
	body, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return pgx.BeginFunc(context.Background(), db, func(tx pgx.Tx) error {
		return ApplyMemberFacts(context.Background(), tx, amqp.Delivery{Body: body})
	})
}

func envelope(member string, at time.Time) *moneyrecapv1.EventEnvelope {
	return &moneyrecapv1.EventEnvelope{MessageId: uuid.New().String(), OccurredAt: timestamppb.New(at), MemberId: member, RoutingKey: "members.updated"}
}

// replica reads the Member's row as "subject zone date code", with "-" for each null, and "" when there is none.
func replica(t *testing.T, db *pgxpool.Pool, member string) string {
	t.Helper()
	var row string
	err := db.QueryRow(context.Background(), `SELECT concat_ws(' ', coalesce(keycloak_subject, '-'), coalesce(timezone, '-'),
		coalesce(to_char(signed_up_on, 'YYYY-MM-DD'), '-'), coalesce(currency_code, '-')) FROM member_facts WHERE member_id = $1`, member).Scan(&row)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal(err)
	}
	return row
}

func forget(t *testing.T, db *pgxpool.Pool, member string) {
	t.Helper()
	t.Cleanup(func() {
		for _, table := range []string{"member_facts", "salary_setting", "idempotency"} {
			if _, err := db.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = $1`, member); err != nil {
				t.Logf("cleanup: %v", err)
			}
		}
	})
}

// The replica follows members' events (§7.2 "Consumes"): members.created shares members.updated's shape, an
// older event arriving late is skipped, an empty fact never clears a stored one, and the subject stays.
func TestApplyMemberFacts(t *testing.T) {
	db := testDB(t)
	id := uuid.New().String()
	forget(t, db, id)
	start := time.Date(2026, 9, 23, 1, 0, 0, 0, time.UTC)
	at := func(minutes int) *moneyrecapv1.EventEnvelope {
		return envelope(id, start.Add(time.Duration(minutes)*time.Minute))
	}
	steps := []struct {
		name string
		msg  proto.Message
		want string
	}{
		{"members.created", &moneyrecapv1.MembersCreated{Envelope: at(0), Facts: &moneyrecapv1.MemberFacts{Email: "m@example.com"}}, "- - - -"},
		{"the first zone", &moneyrecapv1.MembersUpdated{Envelope: at(2), Facts: &moneyrecapv1.MemberFacts{Timezone: "Asia/Jakarta", SignedUpOn: "2026-09-23"}}, "- Asia/Jakarta 2026-09-23 -"},
		{"an older event arriving late", &moneyrecapv1.MembersUpdated{Envelope: at(1), Facts: &moneyrecapv1.MemberFacts{Timezone: "Europe/London", SignedUpOn: "2026-09-22", CurrencyCode: "GBP"}}, "- Asia/Jakarta 2026-09-23 -"},
		{"a Currency", &moneyrecapv1.MembersUpdated{Envelope: at(3), Facts: &moneyrecapv1.MemberFacts{Timezone: "Asia/Jakarta", SignedUpOn: "2026-09-23", CurrencyCode: "IDR"}}, "- Asia/Jakarta 2026-09-23 IDR"},
		{"empty facts", &moneyrecapv1.MembersUpdated{Envelope: at(4), Facts: &moneyrecapv1.MemberFacts{CurrencyCode: "USD"}}, "- Asia/Jakarta 2026-09-23 USD"},
	}
	for _, s := range steps {
		if err := apply(t, db, s.msg); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		if got := replica(t, db, id); got != s.want {
			t.Fatalf("after %s the replica is %q, want %q", s.name, got, s.want)
		}
	}
	if _, err := db.Exec(context.Background(), `UPDATE member_facts SET keycloak_subject = 'owner' WHERE member_id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if err := apply(t, db, &moneyrecapv1.MembersUpdated{Envelope: at(5), Facts: &moneyrecapv1.MemberFacts{Timezone: "Europe/London"}}); err != nil {
		t.Fatal(err)
	}
	if got := replica(t, db, id); got != "owner Europe/London 2026-09-23 USD" {
		t.Fatalf("after a zone change the replica is %q, want the subject kept", got)
	}
	for name, msg := range map[string]proto.Message{
		"no Member":   &moneyrecapv1.MembersUpdated{Envelope: envelope("", start), Facts: &moneyrecapv1.MemberFacts{}},
		"no time":     &moneyrecapv1.MembersUpdated{Envelope: &moneyrecapv1.EventEnvelope{MemberId: id}, Facts: &moneyrecapv1.MemberFacts{}},
		"no facts":    &moneyrecapv1.MembersUpdated{Envelope: at(6)},
		"a bad date":  &moneyrecapv1.MembersUpdated{Envelope: at(6), Facts: &moneyrecapv1.MemberFacts{SignedUpOn: "2026-02-30"}},
		"no envelope": &moneyrecapv1.MembersUpdated{Facts: &moneyrecapv1.MemberFacts{}},
	} {
		if err := apply(t, db, msg); err == nil {
			t.Errorf("%s: applied, want an error so the delivery is retried and then waits for a person", name)
		}
	}
	if err := pgx.BeginFunc(context.Background(), db, func(tx pgx.Tx) error {
		return ApplyMemberFacts(context.Background(), tx, amqp.Delivery{Body: []byte("not a protobuf")})
	}); err == nil {
		t.Error("a body that is not an event: applied, want an error")
	}
}

// Events own the replica's facts: a profile read on a miss stamps the caller's subject and fills only what
// the row lacks, so a read racing a newer event cannot overwrite it (§6.2 step 3, spec #92).
func TestTheProfileReadFillsOnlyMissingFacts(t *testing.T) {
	db := testDB(t)
	id := uuid.New().String()
	forget(t, db, id)
	if err := apply(t, db, &moneyrecapv1.MembersUpdated{Envelope: envelope(id, time.Now()), Facts: &moneyrecapv1.MemberFacts{Timezone: "Asia/Jakarta", SignedUpOn: "2026-09-01"}}); err != nil {
		t.Fatal(err)
	}
	s := &Service{DB: db}
	withMembers(t, s, map[string]*moneyrecapv1.Profile{"owner": {MemberId: id, Timezone: "Europe/London", SignedUpOn: "2026-08-01", CurrencyCode: "IDR"}})
	owner := auth.NewContext(context.Background(), auth.Claims{Subject: "owner", Raw: "owner"})
	if _, err := s.GetSettings(owner, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: id})); err != nil {
		t.Fatal(err)
	}
	if got := replica(t, db, id); got != "owner Asia/Jakarta 2026-09-01 IDR" {
		t.Fatalf("after the profile read the replica is %q, want the events' zone and date, the subject and the Currency filled in", got)
	}
}
