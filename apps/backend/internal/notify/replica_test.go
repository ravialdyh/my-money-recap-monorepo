package notify

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
)

// acks counts what the consumer answered the broker for one delivery.
type acks struct{ acked, rejected int }

func (a *acks) Ack(uint64, bool) error { a.acked++; return nil }

func (a *acks) Nack(uint64, bool, bool) error {
	return errors.New("the consumer nacked: only a reject counts toward five tries")
}

func (a *acks) Reject(uint64, bool) error { a.rejected++; return nil }

// feed hands the consumer one delivery as the broker would and answers whether it was acked, recording its id for
// the cleanup.
type feed struct {
	t   *testing.T
	s   *Service
	ids []string
}

func newFeed(t *testing.T, s *Service) *feed {
	f := &feed{t: t, s: s}
	t.Cleanup(func() {
		if _, err := s.DB.Exec(context.Background(), `DELETE FROM processed_events WHERE message_id = ANY($1)`, f.ids); err != nil {
			t.Logf("cleanup processed_events: %v", err)
		}
	})
	return f
}

func (f *feed) deliver(messageID, key string, msg proto.Message) bool {
	f.t.Helper()
	body, err := proto.Marshal(msg)
	if err != nil {
		f.t.Fatal(err)
	}
	f.ids = append(f.ids, messageID)
	a := &acks{}
	events.Process(f.t.Context(), f.s.DB, amqp.Delivery{Acknowledger: a, MessageId: messageID, RoutingKey: key, Body: body}, ApplyMemberEvent)
	return a.acked == 1 && a.rejected == 0
}

// replicaOf is the Member's copy as "zone email pending", "-" for an empty fact.
func replicaOf(t *testing.T, s *Service, id string) string {
	t.Helper()
	var zone, email string
	var pending bool
	err := s.DB.QueryRow(t.Context(), `SELECT coalesce(timezone, '-'), coalesce(email, '-'), deletion_pending FROM member_facts WHERE member_id = $1`, id).Scan(&zone, &email, &pending)
	if errors.Is(err, pgx.ErrNoRows) {
		return "no row"
	}
	if err != nil {
		t.Fatal(err)
	}
	if pending {
		return zone + " " + email + " pending"
	}
	return zone + " " + email + " clear"
}

var base = time.Date(2026, 9, 29, 8, 0, 0, 0, time.UTC)

func facts(id string, at time.Duration, zone, email string) *moneyrecapv1.MembersUpdated {
	return &moneyrecapv1.MembersUpdated{
		Envelope: &moneyrecapv1.EventEnvelope{MemberId: id, OccurredAt: timestamppb.New(base.Add(at))},
		Facts:    &moneyrecapv1.MemberFacts{Timezone: zone, Email: email, DisplayName: "Rafi", CurrencyCode: "IDR"},
	}
}

func deletion(id string, at time.Duration) *moneyrecapv1.MembersDeletionRequested {
	return &moneyrecapv1.MembersDeletionRequested{Envelope: &moneyrecapv1.EventEnvelope{MemberId: id, OccurredAt: timestamppb.New(base.Add(at))}}
}

// members.created and members.updated give the copy its zone and email, an empty fact never clears one, and the copy
// has nowhere to hold a Display Name; the two deletion events set and clear the flag (spec #169, "The replica").
func TestTheConsumerKeepsTheFactsAndTheFlag(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000c001"
	s := warningService(t, id, nil)
	f := newFeed(t, s)
	steps := []struct {
		key  string
		msg  proto.Message
		want string
	}{
		{"members.created", facts(id, 0, "", "rafi@example.test"), "- rafi@example.test clear"},
		{"members.updated", facts(id, time.Minute, "Asia/Jakarta", ""), "Asia/Jakarta rafi@example.test clear"},
		{"members.updated", facts(id, 2*time.Minute, "Europe/London", "new@example.test"), "Europe/London new@example.test clear"},
		{"members.deletion.requested", deletion(id, 3*time.Minute), "Europe/London new@example.test pending"},
		{"members.deletion.cancelled", deletion(id, 4*time.Minute), "Europe/London new@example.test clear"},
		{"members.deletion.requested", deletion(id, 5*time.Minute), "Europe/London new@example.test pending"},
	}
	for i, step := range steps {
		if !f.deliver(rand.Text(), step.key, step.msg) || replicaOf(t, s, id) != step.want {
			t.Fatalf("step %d, %s: holds %q, want %q and an ack", i, step.key, replicaOf(t, s, id), step.want)
		}
	}
	var displayName int
	if err := s.DB.QueryRow(t.Context(), `SELECT count(*) FROM information_schema.columns WHERE table_name = 'member_facts' AND column_name LIKE '%display%'`).Scan(&displayName); err != nil || displayName != 0 {
		t.Errorf("member_facts has %d Display Name columns (%v), want none: no push prints one", displayName, err)
	}
}

// An older event never overrides a newer one of its kind: a late deletion request after the cancel leaves the flag
// clear, and a late members.updated brings back neither the old zone nor the old email.
func TestALateEventNeverOverridesANewerOne(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000c002"
	s := warningService(t, id, nil)
	f := newFeed(t, s)
	f.deliver(rand.Text(), "members.updated", facts(id, 2*time.Minute, "Europe/London", "new@example.test"))
	f.deliver(rand.Text(), "members.deletion.cancelled", deletion(id, 4*time.Minute))
	if !f.deliver(rand.Text(), "members.deletion.requested", deletion(id, 3*time.Minute)) {
		t.Fatal("the late deletion request was not acked")
	}
	if !f.deliver(rand.Text(), "members.updated", facts(id, time.Minute, "Asia/Jakarta", "old@example.test")) {
		t.Fatal("the late members.updated was not acked")
	}
	if got := replicaOf(t, s, id); got != "Europe/London new@example.test clear" {
		t.Errorf("after both late events the copy holds %q, want the newer zone, email and a clear flag", got)
	}
}

// A second delivery of one message id changes nothing, even carrying other facts, and is acked; a malformed event or
// a key notify.members does not bind is rejected, never nacked, toward its five tries (§4.3).
func TestADuplicateDeliveryChangesNothing(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000c003"
	s := warningService(t, id, nil)
	f := newFeed(t, s)
	once := rand.Text()
	f.deliver(once, "members.updated", facts(id, time.Minute, "Asia/Jakarta", "rafi@example.test"))
	if !f.deliver(once, "members.updated", facts(id, time.Hour, "Europe/London", "other@example.test")) {
		t.Fatal("the duplicate was not acked")
	}
	if got := replicaOf(t, s, id); got != "Asia/Jakarta rafi@example.test clear" {
		t.Errorf("after a duplicate delivery the copy holds %q, want the first delivery's facts", got)
	}
	for name, d := range map[string]struct {
		key string
		msg proto.Message
	}{
		"facts without a Member":  {"members.updated", facts("", time.Minute, "Asia/Jakarta", "x@example.test")},
		"a deletion without time": {"members.deletion.requested", &moneyrecapv1.MembersDeletionRequested{Envelope: &moneyrecapv1.EventEnvelope{MemberId: id}}},
		"an unbound key":          {"ledger.spend.logged", deletion(id, time.Minute)},
		"facts without the facts": {"members.created", &moneyrecapv1.MembersCreated{Envelope: deletion(id, time.Minute).Envelope}},
	} {
		if f.deliver(rand.Text(), d.key, d.msg) {
			t.Errorf("%s: acked, want a reject", name)
		}
	}
}

// The owner check's profile read fills the zone and the email of a row the events never reached (a Member from
// before #176), and a later read leaves a newer event's facts, and a later event older than the read, alone.
func TestTheProfileReadFillsTheCopy(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000c004"
	s := warningService(t, id, map[string]string{"owner": id})
	if _, err := s.DB.Exec(t.Context(), `INSERT INTO member_facts (member_id, keycloak_subject) VALUES ($1, 'owner')`, id); err != nil {
		t.Fatal(err)
	}
	read := func() {
		t.Helper()
		if _, err := s.GetWarningState(as("owner"), connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: id, Month: "2026-09"})); err != nil {
			t.Fatal(err)
		}
	}
	read()
	if got := replicaOf(t, s, id); got != "Asia/Jakarta owner@example.test clear" {
		t.Fatalf("after the owner's read the copy holds %q, want the profile's zone and email", got)
	}
	f := newFeed(t, s)
	f.deliver(rand.Text(), "members.updated", facts(id, -time.Hour, "Europe/London", "old@example.test"))
	if got := replicaOf(t, s, id); got != "Asia/Jakarta owner@example.test clear" {
		t.Errorf("an event older than the read: the copy holds %q, want the read's facts", got)
	}
	if _, err := s.DB.Exec(t.Context(), `UPDATE member_facts SET keycloak_subject = 'earlier', timezone = 'Europe/Paris', facts_at = now() + interval '1 hour' WHERE member_id = $1`, id); err != nil {
		t.Fatal(err)
	}
	read()
	if got := replicaOf(t, s, id); got != "Europe/Paris owner@example.test clear" {
		t.Errorf("a read older than the last event: the copy holds %q, want the event's zone kept", got)
	}
}

// notify's idempotency table is A14's (spec #169, "Idempotency"): a repeated key replays its answer, a key reused for
// another request is invalid_argument, a failed write keeps no key, and keys past 24 hours are swept.
func TestTheIdempotencyTableClaimsAKeyOnce(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000c005"
	s := warningService(t, id, nil)
	t.Cleanup(func() {
		if _, err := s.DB.Exec(context.Background(), `DELETE FROM idempotency WHERE member_id = $1`, id); err != nil {
			t.Logf("cleanup idempotency: %v", err)
		}
	})
	writes := 0
	save := func(key, reminder string, fail bool) (*moneyrecapv1.SavePrefsResponse, error) {
		answer := &moneyrecapv1.SavePrefsResponse{}
		req := &moneyrecapv1.SavePrefsRequest{IdempotencyKey: key, MemberId: id, Prefs: &moneyrecapv1.NotificationPrefs{ReminderTime: reminder}}
		err := s.write(t.Context(), id, key, req, answer, func(pgx.Tx) error {
			if fail {
				return connect.NewError(connect.CodeFailedPrecondition, errors.New("the write refused"))
			}
			writes++
			answer.Prefs = req.Prefs
			return nil
		})
		return answer, err
	}
	if a, err := save("k1", "20:00", false); err != nil || writes != 1 || a.Prefs.GetReminderTime() != "20:00" {
		t.Fatalf("first: %v, %v, %d writes", a, err, writes)
	}
	if a, err := save("k1", "20:00", false); err != nil || writes != 1 || a.Prefs.GetReminderTime() != "20:00" {
		t.Fatalf("repeat: %v, %v, %d writes; want the stored answer and no second write", a, err, writes)
	}
	if _, err := save("k1", "21:00", false); connect.CodeOf(err) != connect.CodeInvalidArgument || writes != 1 {
		t.Fatalf("k1 on another request: %v, %d writes; want invalid_argument and no write", err, writes)
	}
	if _, err := save("k2", "20:00", true); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("a refused write: %v, want its own code", err)
	}
	if _, err := save("k2", "20:00", false); err != nil || writes != 2 {
		t.Fatalf("k2 after its refused write: %v, %d writes; want it written this time", err, writes)
	}
	if _, err := s.DB.Exec(t.Context(), `INSERT INTO idempotency (member_id, idempotency_key, request_hash, created_at) VALUES ($1, 'old', '\x00', now() - interval '25 hours')`, id); err != nil {
		t.Fatal(err)
	}
	if _, err := save("k3", "20:00", false); err != nil {
		t.Fatal(err)
	}
	var old int
	if err := s.DB.QueryRow(t.Context(), `SELECT count(*) FROM idempotency WHERE member_id = $1 AND idempotency_key = 'old'`, id).Scan(&old); err != nil || old != 0 {
		t.Errorf("%d day-old keys after a write (%v), want the sweep to take it", old, err)
	}
}

// A new zone from members' events, and a preferences save, each move the next-fire times to 09:00 on the next 1st and
// the chosen time in the zone now held.
func TestAZoneChangeMovesTheNextTimes(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000d301"
	s := &Service{DB: scheduleDB(t), Members: membersClient(t, map[string]string{"owner": id})}
	_, err := s.SavePrefs(as("owner"), connect.NewRequest(&moneyrecapv1.SavePrefsRequest{
		IdempotencyKey: "k1", MemberId: id, Prefs: &moneyrecapv1.NotificationPrefs{RecapReady: true, DailyReminder: true, ReminderTime: "07:30"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	fires(t, s, id, "Asia/Jakarta", 7, 30)
	f := newFeed(t, s)
	if !f.deliver("zone-d301", "members.updated", &moneyrecapv1.MembersUpdated{
		Envelope: &moneyrecapv1.EventEnvelope{MemberId: id, OccurredAt: timestamppb.New(time.Now().Add(time.Minute))},
		Facts:    &moneyrecapv1.MemberFacts{Timezone: "Europe/London"},
	}) {
		t.Fatal("the zone change was not acked")
	}
	fires(t, s, id, "Europe/London", 7, 30)
}

// fires checks the Member's next Recap ready is 09:00 on the next 1st in zone and the next reminder hour:minute there.
func fires(t *testing.T, s *Service, id, zone string, hour, minute int) {
	t.Helper()
	loc, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	recap, reminder := next(t, s, id, "next_recap_at").In(loc), next(t, s, id, "next_reminder_at").In(loc)
	if recap.Day() != 1 || recap.Hour() != 9 || recap.Minute() != 0 || !recap.After(now) || recap.After(now.AddDate(0, 1, 1)) {
		t.Errorf("in %s the next Recap ready is %v, want 09:00 on the next 1st", zone, recap)
	}
	if reminder.Hour() != hour || reminder.Minute() != minute || !reminder.After(now) || reminder.After(now.Add(24*time.Hour)) {
		t.Errorf("in %s the next reminder is %v, want %02d:%02d within a day", zone, reminder, hour, minute)
	}
}
