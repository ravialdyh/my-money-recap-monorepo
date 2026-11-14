package notify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testdb"
)

// scheduleDB is the notify schema in a schema of its own, dropped at the end, so the stack's notify never ticks the
// test's rows and a test's clock never ticks the stack's.
func scheduleDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testdb.Scratch(t, "NOTIFY_TEST_DATABASE_URL", Migrate)
}

// fakeLedger answers the schedules' two reads: months holds each Member's month list, today each Member's count of
// Spends dated today; a Member in neither is one ledger does not know.
type fakeLedger struct {
	moneyrecapv1connect.UnimplementedLedgerServiceHandler
	months  map[string][]string
	today   map[string]int32
	batches atomic.Int32
}

func (f *fakeLedger) ListMonthsWithData(_ context.Context, req *connect.Request[moneyrecapv1.ListMonthsWithDataRequest]) (*connect.Response[moneyrecapv1.ListMonthsWithDataResponse], error) {
	months, ok := f.months[req.Msg.MemberId]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	return connect.NewResponse(&moneyrecapv1.ListMonthsWithDataResponse{Months: months}), nil
}

func (f *fakeLedger) GetSpentToday(_ context.Context, req *connect.Request[moneyrecapv1.GetSpentTodayRequest]) (*connect.Response[moneyrecapv1.GetSpentTodayResponse], error) {
	f.batches.Add(1)
	res := &moneyrecapv1.GetSpentTodayResponse{}
	for _, id := range req.Msg.MemberIds {
		if count, ok := f.today[id]; ok {
			res.Entries = append(res.Entries, &moneyrecapv1.SpentToday{MemberId: id, SpendCount: count})
		}
	}
	return connect.NewResponse(res), nil
}

// scheduleService ticks against ledger and a fake Apple service over its own schema.
func scheduleService(t *testing.T, ledger *fakeLedger) (*Service, *fakeApple) {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewLedgerServiceHandler(ledger))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	apple, fake, _ := startApple(t)
	return &Service{DB: scheduleDB(t), Ledger: moneyrecapv1connect.NewLedgerServiceClient(srv.Client(), srv.URL), Apple: apple}, fake
}

// scheduled is a Member in zone with one device named after the id, the reminder at reminder ("" keeps the defaults),
// and next-fire times set from now.
func scheduled(t *testing.T, s *Service, id, zone, reminder string, now time.Time) {
	t.Helper()
	if _, err := s.DB.Exec(t.Context(), `INSERT INTO member_facts (member_id, timezone) VALUES ($1, $2)`, id, zone); err != nil {
		t.Fatal(err)
	}
	if reminder != "" {
		if _, err := s.DB.Exec(t.Context(), `INSERT INTO notification_preference VALUES ($1, true, true, $2, true)`, id, reminder); err != nil {
			t.Fatal(err)
		}
	}
	devices(t, s, id, deviceOf(id))
	if err := reschedule(t.Context(), s.DB, id, now); err != nil {
		t.Fatal(err)
	}
}

func deviceOf(id string) string { return strings.ReplaceAll(id, "-", "") }

// next is the Member's next-fire time in column, the zero time when it is NULL.
func next(t *testing.T, s *Service, id, column string) time.Time {
	t.Helper()
	var at *time.Time
	if err := s.DB.QueryRow(t.Context(), `SELECT `+column+` FROM member_facts WHERE member_id = $1`, id).Scan(&at); err != nil {
		t.Fatal(err)
	}
	if at == nil {
		return time.Time{}
	}
	return at.UTC()
}

// alerts is each device's alert bodies, in the order Apple received them.
func alerts(t *testing.T, fake *fakeApple) map[string][]string {
	t.Helper()
	got := map[string][]string{}
	for _, p := range fake.received() {
		var payload struct {
			Aps struct {
				Alert struct{ Body string } `json:"alert"`
			} `json:"aps"`
		}
		if err := json.Unmarshal([]byte(p.body), &payload); err != nil {
			t.Fatal(err)
		}
		device := strings.TrimPrefix(p.path, "/3/device/")
		got[device] = append(got[device], payload.Aps.Alert.Body)
	}
	return got
}

func utc(s string) time.Time {
	at, err := time.Parse(time.DateTime, s)
	if err != nil {
		panic(err)
	}
	return at
}

// Recap ready fires at 09:00 on the 1st in each Member's zone, 02:00 UTC in Jakarta and 08:00 UTC in a London summer,
// at 09:00 UTC once London's clocks go back, and a second run, or a replay of the same fire, sends nothing more.
func TestRecapReadyAtNineOnTheFirstInTheMembersZone(t *testing.T) {
	london, jakarta := "00000000-0000-4000-8000-00000000d001", "00000000-0000-4000-8000-00000000d002"
	ledger := &fakeLedger{months: map[string][]string{london: {"2026-10", "2026-09", "2026-08"}, jakarta: {"2026-09", "2026-08"}}}
	s, fake := scheduleService(t, ledger)
	start := utc("2026-08-15 12:00:00")
	scheduled(t, s, london, "Europe/London", "", start)
	scheduled(t, s, jakarta, "Asia/Jakarta", "", start)
	if l, j := next(t, s, london, "next_recap_at"), next(t, s, jakarta, "next_recap_at"); !l.Equal(utc("2026-09-01 08:00:00")) || !j.Equal(utc("2026-09-01 02:00:00")) {
		t.Fatalf("next Recap ready: London %v, Jakarta %v, want 08:00 and 02:00 UTC on 1 September", l, j)
	}
	august := "Your August recap is ready. See how you did."
	for _, step := range []struct {
		at             string
		heardInLondon  []string
		heardInJakarta []string
	}{
		{"2026-09-01 01:59:00", nil, nil},
		{"2026-09-01 02:00:00", nil, []string{august}},
		{"2026-09-01 02:00:30", nil, []string{august}},
		{"2026-09-01 07:59:00", nil, []string{august}},
		{"2026-09-01 08:00:00", []string{august}, []string{august}},
		{"2026-09-01 08:01:00", []string{august}, []string{august}},
	} {
		s.tick(t.Context(), utc(step.at))
		got := alerts(t, fake)
		if strings.Join(got[deviceOf(london)], "|") != strings.Join(step.heardInLondon, "|") || strings.Join(got[deviceOf(jakarta)], "|") != strings.Join(step.heardInJakarta, "|") {
			t.Fatalf("at %s UTC London heard %q and Jakarta %q, want %q and %q", step.at, got[deviceOf(london)], got[deviceOf(jakarta)], step.heardInLondon, step.heardInJakarta)
		}
	}
	if l := next(t, s, london, "next_recap_at"); !l.Equal(utc("2026-10-01 08:00:00")) {
		t.Fatalf("after its fire London's next Recap ready is %v, want 08:00 UTC on 1 October", l)
	}
	if _, err := s.DB.Exec(t.Context(), `UPDATE member_facts SET next_recap_at = '2026-09-01 08:00:00+00' WHERE member_id = $1`, london); err != nil {
		t.Fatal(err)
	}
	s.tick(t.Context(), utc("2026-09-01 08:02:00"))
	if got := alerts(t, fake)[deviceOf(london)]; len(got) != 1 {
		t.Fatalf("a replay of September's fire, as after a crash before the advance, sent %q, want nothing more", got)
	}
	s.tick(t.Context(), utc("2026-10-01 08:00:00"))
	if got := alerts(t, fake)[deviceOf(london)]; len(got) != 2 || got[1] != "Your September recap is ready. See how you did." {
		t.Fatalf("on 1 October London heard %q, want September's recap", got)
	}
	if l := next(t, s, london, "next_recap_at"); !l.Equal(utc("2026-11-01 09:00:00")) {
		t.Errorf("after the clocks go back London's next Recap ready is %v, want 09:00 UTC on 1 November", l)
	}
}

// Recap ready sends only when last month is on ledger's list, a partial sign-up month included, and never to a Member
// ledger does not know, whose switch is off or whose deletion is pending; each still moves to the next 1st.
func TestRecapReadyNeedsLastMonthsData(t *testing.T) {
	empty, partial, unknown, off, leaving := "00000000-0000-4000-8000-00000000d101", "00000000-0000-4000-8000-00000000d102",
		"00000000-0000-4000-8000-00000000d103", "00000000-0000-4000-8000-00000000d104", "00000000-0000-4000-8000-00000000d105"
	both := []string{"2026-09", "2026-08"}
	ledger := &fakeLedger{months: map[string][]string{empty: {"2026-09"}, partial: both, off: both, leaving: both}}
	s, fake := scheduleService(t, ledger)
	start := utc("2026-08-15 12:00:00")
	for _, id := range []string{empty, partial, unknown, off, leaving} {
		scheduled(t, s, id, "Asia/Jakarta", "", start)
	}
	if _, err := s.DB.Exec(t.Context(), `INSERT INTO notification_preference VALUES ($1, false, false, '20:00', true)`, off); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.Exec(t.Context(), `UPDATE member_facts SET deletion_pending = true WHERE member_id = $1`, leaving); err != nil {
		t.Fatal(err)
	}
	s.tick(t.Context(), utc("2026-09-01 02:00:00"))
	got := alerts(t, fake)
	if len(got) != 1 || len(got[deviceOf(partial)]) != 1 {
		t.Errorf("Apple heard %q, want August's recap for the partial month alone", got)
	}
	for _, id := range []string{empty, unknown, leaving} {
		if at := next(t, s, id, "next_recap_at"); !at.Equal(utc("2026-10-01 02:00:00")) {
			t.Errorf("%s: next Recap ready %v, want 1 October", id, at)
		}
	}
	var claimed int
	if err := s.DB.QueryRow(t.Context(), `SELECT count(*) FROM sent_log`).Scan(&claimed); err != nil || claimed != 1 {
		t.Errorf("sent_log holds %d rows (%v), want the one push sent", claimed, err)
	}
}

// The reminder asks ledger once for the batch and sends only on no Spend dated today, never with the switch off or a
// deletion pending, then moves each Member to the next day; a second run, or a replay, sends nothing more.
func TestTheReminderSkipsADayWithASpend(t *testing.T) {
	spent, idle, off, leaving := "00000000-0000-4000-8000-00000000d201", "00000000-0000-4000-8000-00000000d202",
		"00000000-0000-4000-8000-00000000d203", "00000000-0000-4000-8000-00000000d204"
	ledger := &fakeLedger{today: map[string]int32{spent: 1, idle: 0, off: 0, leaving: 0}}
	s, fake := scheduleService(t, ledger)
	start := utc("2026-09-30 00:00:00")
	for _, id := range []string{spent, idle, leaving} {
		scheduled(t, s, id, "Europe/London", "20:00", start)
	}
	scheduled(t, s, off, "Europe/London", "", start)
	if _, err := s.DB.Exec(t.Context(), `UPDATE member_facts SET deletion_pending = true WHERE member_id = $1`, leaving); err != nil {
		t.Fatal(err)
	}
	if at := next(t, s, off, "next_reminder_at"); !at.IsZero() {
		t.Errorf("with the reminder off the next reminder is %v, want none", at)
	}
	s.tick(t.Context(), utc("2026-09-30 18:59:00"))
	s.tick(t.Context(), utc("2026-09-30 19:00:00"))
	s.tick(t.Context(), utc("2026-09-30 19:01:00"))
	got := alerts(t, fake)
	if len(got) != 1 || strings.Join(got[deviceOf(idle)], "|") != "Log today's spends before you forget." {
		t.Errorf("Apple heard %q, want the reminder once, to the Member with no Spend today", got)
	}
	if n := ledger.batches.Load(); n != 1 {
		t.Errorf("ledger was asked %d times, want once for the batch", n)
	}
	for _, id := range []string{spent, idle, leaving} {
		if at := next(t, s, id, "next_reminder_at"); !at.Equal(utc("2026-10-01 19:00:00")) {
			t.Errorf("%s: next reminder %v, want 20:00 London on 1 October", id, at)
		}
	}
	if _, err := s.DB.Exec(t.Context(), `UPDATE member_facts SET next_reminder_at = '2026-09-30 19:00:00+00' WHERE member_id = $1`, idle); err != nil {
		t.Fatal(err)
	}
	s.tick(t.Context(), utc("2026-09-30 19:02:00"))
	if got := alerts(t, fake)[deviceOf(idle)]; len(got) != 1 {
		t.Errorf("a replay of the same day's fire sent %q, want nothing more", got)
	}
}
