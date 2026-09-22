package members

import (
	"context"
	"crypto/rand"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// staged is one outbox row: its id, routing key and the event it carries, which members.created and
// members.updated share the shape of.
type staged struct {
	id, key string
	event   *moneyrecapv1.MembersUpdated
}

func outbox(t *testing.T, db *pgxpool.Pool, member string) []staged {
	t.Helper()
	rows, err := db.Query(context.Background(), `SELECT id::text, routing_key, payload FROM outbox WHERE member_id = $1 ORDER BY created_at, id`, member)
	if err != nil {
		t.Fatal(err)
	}
	found, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (staged, error) {
		var s staged
		var payload []byte
		s.event = &moneyrecapv1.MembersUpdated{}
		if err := r.Scan(&s.id, &s.key, &payload); err != nil {
			return s, err
		}
		return s, proto.Unmarshal(payload, s.event)
	})
	if err != nil {
		t.Fatal(err)
	}
	return found
}

// Each change to a Member's facts stages exactly one event carrying all of them, in the change's own
// transaction; a write that changes none stages nothing: an Avatar, a step, a repeated answer, a replay.
func TestFactChangesStageOneEventEach(t *testing.T) {
	db := testDB(t)
	subject := "outbox-" + rand.Text()
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject})
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM outbox WHERE member_id IN (SELECT id FROM member WHERE keycloak_subject = $1)`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	m, err := Upsert(ctx, db, subject, "first@example.com")
	if err != nil {
		t.Fatal(err)
	}
	s := &Service{DB: db}
	seen := 0
	want := func(step, key string, facts *moneyrecapv1.MemberFacts) {
		t.Helper()
		found := outbox(t, db, m.ID)
		fresh := found[seen:]
		seen = len(found)
		if key == "" {
			if len(fresh) != 0 {
				t.Fatalf("%s staged %d events, want none", step, len(fresh))
			}
			return
		}
		if len(fresh) != 1 {
			t.Fatalf("%s staged %d events, want 1", step, len(fresh))
		}
		env := fresh[0].event.GetEnvelope()
		if fresh[0].key != key || env.GetRoutingKey() != key || env.GetMessageId() != fresh[0].id || env.GetMemberId() != m.ID || env.GetOccurredAt() == nil {
			t.Fatalf("%s staged %s carrying %v", step, fresh[0].key, env)
		}
		if !proto.Equal(fresh[0].event.GetFacts(), facts) {
			t.Fatalf("%s staged facts %v, want %v", step, fresh[0].event.GetFacts(), facts)
		}
	}
	zone := func(key, name string) string {
		t.Helper()
		res, err := s.SetTimezone(ctx, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: key, MemberId: m.ID, Timezone: name}))
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg.Profile.SignedUpOn
	}
	save := func(key string, req *moneyrecapv1.SaveProfileRequest) {
		t.Helper()
		req.IdempotencyKey, req.MemberId = key, m.ID
		if _, err := s.SaveProfile(ctx, connect.NewRequest(req)); err != nil {
			t.Fatal(err)
		}
	}
	facts := &moneyrecapv1.MemberFacts{Email: "first@example.com"}
	want("first sight", "members.created", facts)
	for _, email := range []string{"first@example.com", ""} {
		if _, err := Upsert(ctx, db, subject, email); err != nil {
			t.Fatal(err)
		}
		want("a sign-in with the same or no email", "", nil)
	}
	if _, err := Upsert(ctx, db, subject, "second@example.com"); err != nil {
		t.Fatal(err)
	}
	facts.Email = "second@example.com"
	want("a new email", "members.updated", facts)
	facts.Timezone, facts.SignedUpOn = "Asia/Jakarta", zone("t1", "Asia/Jakarta")
	want("the first zone", "members.updated", facts)
	zone("t2", "Asia/Jakarta")
	want("the same zone again", "", nil)
	save("p1", &moneyrecapv1.SaveProfileRequest{AvatarKey: proto.String("fox"), OnboardingStep: proto.Int32(1)})
	want("an Avatar-only save", "", nil)
	save("p2", &moneyrecapv1.SaveProfileRequest{DisplayName: proto.String("Maya"), AvatarKey: proto.String("otter"), OnboardingStep: proto.Int32(1)})
	facts.DisplayName = "Maya"
	want("a Display Name", "members.updated", facts)
	save("p3", &moneyrecapv1.SaveProfileRequest{DisplayName: proto.String("Maya"), OnboardingStep: proto.Int32(2)})
	want("the same name and a later step", "", nil)
	save("p4", &moneyrecapv1.SaveProfileRequest{CurrencyCode: proto.String("IDR"), OnboardingStep: proto.Int32(2)})
	facts.CurrencyCode = "IDR"
	want("a Currency", "members.updated", facts)
	save("p4", &moneyrecapv1.SaveProfileRequest{CurrencyCode: proto.String("IDR"), OnboardingStep: proto.Int32(2)})
	want("a replayed key", "", nil)
	facts.Timezone = "Europe/London"
	if zone("t3", "Europe/London") != facts.SignedUpOn {
		t.Fatal("a later zone moved the sign-up date")
	}
	want("a new zone", "members.updated", facts)
}
