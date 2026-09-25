package ledger

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// spendsOf answers the Member's ledger rows as a test sets them up: a Service on Jakarta's 20 September 2026, a
// Member since July on a 3,000,000 Salary with a 20% Target (a 2,400,000 line), and a Spend logger.
func spendsOf(t *testing.T, id, subject string) (context.Context, *Service, func(key, amount, date string) *moneyrecapv1.LogSpendResponse) {
	t.Helper()
	db := testDB(t)
	ctx := seed(t, db, id, subject, "Asia/Jakarta", "2026-07-01", "IDR",
		[]*moneyrecapv1.SalarySetting{{Amount: idr("3000000"), OriginalAmount: idr("3000000"), Payday: "25", EffectiveFrom: "2026-07"}},
		[]*moneyrecapv1.SavingTargetSetting{{Percent: 20, EffectiveFrom: "2026-07"}})
	s := &Service{DB: db, Now: func() time.Time { return at(t, "2026-09-20", "Asia/Jakarta") }}
	withMembers(t, s, nil)
	log := func(key, amount, date string) *moneyrecapv1.LogSpendResponse {
		t.Helper()
		res, err := s.LogSpend(ctx, connect.NewRequest(&moneyrecapv1.LogSpendRequest{
			IdempotencyKey: key, MemberId: id, Amount: idr(amount), Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: date,
		}))
		if err != nil {
			t.Fatalf("a Spend of %s on %s: %v", amount, date, err)
		}
		return res.Msg
	}
	return ctx, s, log
}

// An edit crosses the line as a log does, replays a key and refuses it reused; moving a Spend to last month changes
// both months; a date older than 60 days may stay but a Spend never moves to or from one; another Member's is not_found.
func TestEditSpend(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000a801"
	ctx, s, log := spendsOf(t, id, "editor")
	edit := func(key, spendID, amount, date string) (*moneyrecapv1.EditSpendResponse, error) {
		res, err := s.EditSpend(ctx, connect.NewRequest(&moneyrecapv1.EditSpendRequest{
			IdempotencyKey: key, MemberId: id, SpendId: spendID, Amount: idr(amount), Category: moneyrecapv1.Category_CATEGORY_BILLS, Date: date, Note: "Listrik",
		}))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	month := func(m string) *moneyrecapv1.MonthSummary {
		t.Helper()
		res, err := s.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: m}))
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg.Summary
	}

	large := log("l1", "2000000", "2026-09-20").Spend
	small := log("l2", "300000", "2026-09-10").Spend
	crossed, err := edit("e1", small.Id, "500000", "2026-09-10")
	if err != nil || !crossed.CrossedLine || crossed.Month.Month != "2026-09" || crossed.Month.Spent.Amount != "2500000" ||
		crossed.Spend.Note != "Listrik" || crossed.Spend.Category != moneyrecapv1.Category_CATEGORY_BILLS || !proto.Equal(crossed.Spend.CreatedAt, small.CreatedAt) {
		t.Fatalf("300000 edited to 500000 past the line: %v, %v", crossed, err)
	}
	if again, err := edit("e1", small.Id, "500000", "2026-09-10"); err != nil || !proto.Equal(again, crossed) {
		t.Errorf("e1 again: %v, %v, want the first answer", again, err)
	}
	if _, err := edit("e1", small.Id, "500001", "2026-09-10"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("e1 reused for another amount: %v, want invalid_argument", err)
	}
	if under, err := edit("e2", small.Id, "100000", "2026-09-10"); err != nil || under.CrossedLine || under.Month.TargetWarning != nil {
		t.Errorf("back under the line: %v, %v, want no warning and crossed_line false", under, err)
	}

	moved, err := edit("e3", large.Id, "2000000", "2026-08-20")
	if err != nil || moved.CrossedLine || moved.Month.Month != "2026-08" || moved.Month.Spent.Amount != "2000000" || moved.Spend.Date != "2026-08-20" {
		t.Fatalf("the 2000000 Spend moved to August: %v, %v, want August's month with it", moved, err)
	}
	if sep := month("2026-09"); sep.Spent.Amount != "100000" {
		t.Errorf("September after the move: Spent %v, want only the 100000 left", sep.Spent)
	}

	seedSpends(t, s.DB, id, []*moneyrecapv1.Spend{{Amount: idr("70000"), OriginalAmount: idr("70000"), Category: 1, Date: "2026-07-15"}})
	var old string
	if err := s.DB.QueryRow(ctx, `SELECT id::text FROM spend WHERE member_id = $1 AND spent_on = '2026-07-15'`, id).Scan(&old); err != nil {
		t.Fatal(err)
	}
	if kept, err := edit("e4", old, "80000", "2026-07-15"); err != nil || kept.Month.Month != "2026-07" || kept.Month.Spent.Amount != "80000" {
		t.Errorf("a July Spend edited on its own date: %v, %v, want it kept in July", kept, err)
	}
	for name, e := range map[string]struct{ spend, date string }{
		"an old Spend moved into the window":   {old, "2026-09-01"},
		"a recent Spend moved past 60 days":    {small.Id, "2026-07-15"},
		"a recent Spend moved to tomorrow":     {small.Id, "2026-09-21"},
		"a date that is not YYYY-MM-DD":        {small.Id, "20 Sep"},
		"an old Spend moved to the 60th day":   {old, "2026-07-23"},
		"a recent Spend moved to the 61st day": {small.Id, "2026-07-22"},
	} {
		if _, err := edit("bad "+name, e.spend, "1", e.date); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", name, err)
		}
	}
	if _, err := edit("", small.Id, "1", "2026-09-10"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no key: %v, want invalid_argument", err)
	}
	if _, err := edit("e5", small.Id, "0", "2026-09-10"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("an amount of 0: %v, want invalid_argument", err)
	}

	otherID := "00000000-0000-4000-8000-00000000a802"
	otherCtx, _, otherLog := spendsOf(t, otherID, "other-editor")
	theirs := otherLog("o1", "1000", "2026-09-20").Spend
	for name, spendID := range map[string]string{"another Member's Spend": theirs.Id, "no such Spend": "00000000-0000-4000-8000-000000000000", "not an id": "x"} {
		if _, err := edit("nf "+name, spendID, "1", "2026-09-20"); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("%s: %v, want not_found", name, err)
		}
	}
	stranger := auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})
	if _, err := s.EditSpend(stranger, connect.NewRequest(&moneyrecapv1.EditSpendRequest{
		IdempotencyKey: "s1", MemberId: otherID, SpendId: theirs.Id, Amount: idr("1"), Category: 1, Date: "2026-09-20",
	})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another token: %v, want not_found", err)
	}
	res, err := s.ListSpends(otherCtx, connect.NewRequest(&moneyrecapv1.ListSpendsRequest{MemberId: otherID, Month: "2026-09"}))
	if err != nil || len(res.Msg.Spends) != 1 || res.Msg.Spends[0].Amount.Amount != "1000" {
		t.Errorf("the other Member's September: %v, %v, want their Spend untouched", res, err)
	}
}

// A delete leaves every figure as if the Spend had never been logged, answers the month it left with crossed_line
// false, replays its key; a deleted Spend, and another Member's, is not_found.
func TestDeleteSpend(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000a803"
	ctx, s, log := spendsOf(t, id, "deleter")
	del := func(key, spendID string) (*moneyrecapv1.DeleteSpendResponse, error) {
		res, err := s.DeleteSpend(ctx, connect.NewRequest(&moneyrecapv1.DeleteSpendRequest{IdempotencyKey: key, MemberId: id, SpendId: spendID}))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	figures := func(m string) (*moneyrecapv1.GetMonthSummaryResponse, *moneyrecapv1.ListSpendsResponse, *moneyrecapv1.GetRecapResponse) {
		t.Helper()
		summary, err := s.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: m}))
		if err != nil {
			t.Fatal(err)
		}
		list, err := s.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.ListSpendsRequest{MemberId: id, Month: m}))
		if err != nil {
			t.Fatal(err)
		}
		recap, err := s.GetRecap(ctx, connect.NewRequest(&moneyrecapv1.GetRecapRequest{MemberId: id, Month: m}))
		if err != nil {
			t.Fatal(err)
		}
		return summary.Msg, list.Msg, recap.Msg
	}

	log("l1", "2000000", "2026-09-05")
	log("l2", "150000", "2026-08-12")
	sep, sepList, sepRecap := figures("2026-09")
	aug, augList, augRecap := figures("2026-08")
	now := log("l3", "600000", "2026-09-19")
	if !now.CrossedLine {
		t.Fatalf("600000 over a 2000000 September: %v, want it to cross the 2400000 line", now)
	}
	last := log("l4", "90000", "2026-08-30")

	gone, err := del("d1", now.Spend.Id)
	if err != nil || gone.CrossedLine || gone.Month.Month != "2026-09" || gone.Month.Spent.Amount != "2000000" || !proto.Equal(gone.Spend, now.Spend) {
		t.Fatalf("deleting September's crossing Spend: %v, %v, want it as it was and September back at 2000000", gone, err)
	}
	if again, err := del("d1", now.Spend.Id); err != nil || !proto.Equal(again, gone) {
		t.Errorf("d1 again: %v, %v, want the first answer", again, err)
	}
	if _, err := del("d2", now.Spend.Id); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("a deleted Spend under a new key: %v, want not_found", err)
	}
	if left, err := del("d3", last.Spend.Id); err != nil || left.Month.Month != "2026-08" || left.Month.Spent.Amount != "150000" {
		t.Errorf("deleting an August Spend: %v, %v, want August's month without it", left, err)
	}
	for m, before := range map[string][]proto.Message{"2026-09": {sep, sepList, sepRecap}, "2026-08": {aug, augList, augRecap}} {
		summary, list, recap := figures(m)
		for i, after := range []proto.Message{summary, list, recap} {
			if !proto.Equal(before[i], after) {
				t.Errorf("%s after the deletes: %v, want every figure as before the deleted Spends: %v", m, after, before[i])
			}
		}
	}

	otherID := "00000000-0000-4000-8000-00000000a804"
	_, _, otherLog := spendsOf(t, otherID, "other-deleter")
	theirs := otherLog("o1", "1000", "2026-09-20").Spend
	if _, err := del("d4", theirs.Id); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another Member's Spend: %v, want not_found", err)
	}
	stranger := auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})
	if _, err := s.DeleteSpend(stranger, connect.NewRequest(&moneyrecapv1.DeleteSpendRequest{IdempotencyKey: "s1", MemberId: otherID, SpendId: theirs.Id})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another token: %v, want not_found", err)
	}
	if _, err := del("", theirs.Id); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no key: %v, want invalid_argument", err)
	}
}

// Every Spend write stages its event in its own transaction, carrying the answer's Spend, month and crossed_line
// (ARCHITECTURE §4.3, A4); a replayed key stages nothing more.
func TestEverySpendWriteStagesItsEvent(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000a805"
	ctx, s, log := spendsOf(t, id, "announcer")
	type staged struct {
		key     string
		spend   *moneyrecapv1.Spend
		month   *moneyrecapv1.MonthSummary
		crossed bool
	}
	under := log("l1", "2000000", "2026-09-05")
	over := log("l2", "500000", "2026-09-10")
	log("l2", "500000", "2026-09-10")
	edited, err := s.EditSpend(ctx, connect.NewRequest(&moneyrecapv1.EditSpendRequest{
		IdempotencyKey: "e1", MemberId: id, SpendId: under.Spend.Id, Amount: idr("2100000"), Category: 1, Date: "2026-09-05",
	}))
	if err != nil {
		t.Fatal(err)
	}
	gone, err := s.DeleteSpend(ctx, connect.NewRequest(&moneyrecapv1.DeleteSpendRequest{IdempotencyKey: "d1", MemberId: id, SpendId: over.Spend.Id}))
	if err != nil {
		t.Fatal(err)
	}
	want := []staged{
		{"ledger.spend.logged", under.Spend, under.Month, false},
		{"ledger.spend.logged", over.Spend, over.Month, true},
		{"ledger.spend.edited", edited.Msg.Spend, edited.Msg.Month, false},
		{"ledger.spend.deleted", gone.Msg.Spend, gone.Msg.Month, false},
	}
	rows, err := s.DB.Query(ctx, `SELECT id::text, routing_key, payload FROM outbox WHERE member_id = $1 ORDER BY created_at`, id)
	if err != nil {
		t.Fatal(err)
	}
	got, err := pgx.CollectRows(rows, pgx.RowToStructByPos[struct {
		ID, Key string
		Payload []byte
	}])
	if err != nil || len(got) != len(want) {
		t.Fatalf("the outbox holds %d rows (%v), want %d", len(got), err, len(want))
	}
	for i, w := range want {
		var e moneyrecapv1.LedgerSpendLogged
		if err := proto.Unmarshal(got[i].Payload, &e); err != nil {
			t.Fatal(err)
		}
		env := e.GetEnvelope()
		if got[i].Key != w.key || env.GetRoutingKey() != w.key || env.GetMessageId() != got[i].ID || env.GetMemberId() != id ||
			!proto.Equal(e.Spend, w.spend) || !proto.Equal(e.Month, w.month) || e.CrossedLine != w.crossed {
			t.Errorf("outbox row %d: %s %v, want %s carrying the answer %v %v crossed_line %v", i, got[i].Key, &e, w.key, w.spend, w.month, w.crossed)
		}
	}
}
