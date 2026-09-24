package ledger

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// LogSpend validates in the Member's zone, replays a key and refuses it reused, crosses the line once and only in
// the current month, and a token that does not own the Member reaches nothing (§4.2 point 3).
func TestLogSpend(t *testing.T) {
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a501"
	ctx := seed(t, db, id, "spender", "Asia/Jakarta", "2026-07-01", "IDR",
		[]*moneyrecapv1.SalarySetting{{Amount: idr("3000000"), OriginalAmount: idr("3000000"), Payday: "25", EffectiveFrom: "2026-07"}},
		[]*moneyrecapv1.SavingTargetSetting{{Percent: 20, EffectiveFrom: "2026-07"}})
	clock := time.Date(2026, 9, 19, 18, 0, 0, 0, time.UTC)
	s := &Service{DB: db, Now: func() time.Time { return clock }}
	log := func(key, amount, date, note string) (*moneyrecapv1.LogSpendResponse, error) {
		res, err := s.LogSpend(ctx, connect.NewRequest(&moneyrecapv1.LogSpendRequest{
			IdempotencyKey: key, MemberId: id, Amount: idr(amount), Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: date, Note: note,
		}))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}

	composed := strings.Repeat("e\u0301", 80)
	first, err := log("k1", "2000000", "2026-09-20", composed)
	if err != nil || first.CrossedLine || first.Month.Month != "2026-09" || first.Month.Spent.Amount != "2000000" || first.Spend.Id == "" {
		t.Fatalf("a Spend on Jakarta's 20 September at 18:00 UTC: %v, %v", first, err)
	}
	again, err := log("k1", "2000000", "2026-09-20", composed)
	if err != nil || again.Spend.Id != first.Spend.Id {
		t.Fatalf("k1 again: %v, %v, want the first answer", again, err)
	}
	if _, err := log("k1", "2000001", "2026-09-20", composed); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("k1 reused for another amount: %v, want invalid_argument", err)
	}
	backdated, err := log("k2", "2500000", "2026-08-10", "")
	if err != nil || backdated.CrossedLine || backdated.Month.Month != "2026-08" || backdated.Month.TargetWarning == nil {
		t.Fatalf("a Spend past August's line: %v, %v, want August over the line and crossed_line false", backdated, err)
	}
	crossed, err := log("k3", "400001", "2026-09-20", "")
	if err != nil || !crossed.CrossedLine || crossed.Month.TargetWarning.GetMessage() != "You've spent Rp2.400.001 of the Rp2.400.000 you meant to spend this month." {
		t.Fatalf("the Spend past September's line: %v, %v", crossed, err)
	}
	if past, err := log("k4", "1", "2026-09-20", ""); err != nil || past.CrossedLine {
		t.Errorf("a Spend already past the line: %v, %v, want crossed_line false", past, err)
	}
	for name, req := range map[string]*moneyrecapv1.LogSpendRequest{
		"no key":            {MemberId: id, Amount: idr("1"), Category: 1, Date: "2026-09-20"},
		"zero":              {IdempotencyKey: "b1", MemberId: id, Amount: idr("0"), Category: 1, Date: "2026-09-20"},
		"twelve digits":     {IdempotencyKey: "b2", MemberId: id, Amount: idr("100000000000"), Category: 1, Date: "2026-09-20"},
		"a rupiah cent":     {IdempotencyKey: "b3", MemberId: id, Amount: idr("1.5"), Category: 1, Date: "2026-09-20"},
		"no category":       {IdempotencyKey: "b4", MemberId: id, Amount: idr("1"), Date: "2026-09-20"},
		"a ninth category":  {IdempotencyKey: "b5", MemberId: id, Amount: idr("1"), Category: 9, Date: "2026-09-20"},
		"tomorrow":          {IdempotencyKey: "b6", MemberId: id, Amount: idr("1"), Category: 1, Date: "2026-09-21"},
		"the 61st day back": {IdempotencyKey: "b7", MemberId: id, Amount: idr("1"), Category: 1, Date: "2026-07-22"},
		"no date":           {IdempotencyKey: "b8", MemberId: id, Amount: idr("1"), Category: 1},
		"an 81st character": {IdempotencyKey: "b9", MemberId: id, Amount: idr("1"), Category: 1, Date: "2026-09-20", Note: composed + "x"},
	} {
		if _, err := s.LogSpend(ctx, connect.NewRequest(req)); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", name, err)
		}
	}
	if _, err := log("b10", "1", "2026-07-23", ""); err != nil {
		t.Errorf("60 days back including today: %v", err)
	}

	joiner := "00000000-0000-4000-8000-00000000a502"
	joined := seed(t, db, joiner, "joiner", "Asia/Jakarta", "2026-09-10", "IDR", nil, nil)
	early := func(key, date string) error {
		_, err := s.LogSpend(joined, connect.NewRequest(&moneyrecapv1.LogSpendRequest{
			IdempotencyKey: key, MemberId: joiner, Amount: idr("1"), Category: 1, Date: date,
		}))
		return err
	}
	if err := early("j1", "2026-08-31"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("a Spend before the sign-up month (D72): %v, want invalid_argument", err)
	}
	if err := early("j2", "2026-09-01"); err != nil {
		t.Errorf("the 1st of the sign-up month, before the sign-up day (D72): %v", err)
	}
	month, err := s.GetMonthSummary(joined, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: joiner, Month: "2026-09"}))
	if err != nil || month.Msg.EarliestSpendDate != "2026-09-01" || month.Msg.Summary.Spent.Amount != "1" {
		t.Errorf("the joiner's September: %v, %v, want 2026-09-01 as the earliest date and the Spend counted", month, err)
	}
	if month, err := s.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: "2026-09"})); err != nil || month.Msg.EarliestSpendDate != "2026-07-23" {
		t.Errorf("a Member since July: %v, %v, want 59 days back, 2026-07-23", month, err)
	}
	withMembers(t, s, nil)
	stranger := auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})
	if _, err := s.LogSpend(stranger, connect.NewRequest(&moneyrecapv1.LogSpendRequest{
		IdempotencyKey: "s1", MemberId: id, Amount: idr("1"), Category: 1, Date: "2026-09-20",
	})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another token: %v, want not_found", err)
	}
}

// A past month with Spends and no Salary is partial (§8.11); one that missed its target says by how much (§8.7).
func TestPastMonthsWithSpends(t *testing.T) {
	c, _ := currencyOf("GBP")
	today := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	l := ledgerOf{cash: c, signedUp: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		salaries: []salaryRow{{amount: big.NewRat(2400, 1), payday: "25", from: month{2026, time.September}}},
		targets:  []targetRow{{percent: 20, from: month{2026, time.September}}},
		spends: []spendRow{
			{amount: big.NewRat(50, 1), date: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)},
			{amount: big.NewRat(2000, 1), date: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
		}}
	if !(ledgerOf{cash: c, spends: l.spends}).settings(today).HasAmountsToConvert {
		t.Error("Spends and no Salary: nothing to convert, want something (§8.9 step 2)")
	}
	if aug := l.summary(month{2026, time.August}, today); !aug.Partial || aug.StreakDay != 28 {
		t.Errorf("August: %v, want partial with a 28-day Streak", aug)
	}
	sep := l.summary(month{2026, time.September}, today)
	if sep.SavingTarget.ProgressSentence != "£80 short of your target." || sep.TargetWarning.GetMessage() != "You've passed your saving line. £80 over." {
		t.Errorf("September: %v", sep)
	}
}
