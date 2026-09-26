package ledger

import (
	"context"
	"math/big"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// Home's list is every past month with a Spend or Money In, newest first, by Archetype or "Partial", the only past
// months that answer; the current month is the preview; another token reads none (PRD §6.8, §8.11, §4.2).
func TestRecapsAndHomesList(t *testing.T) {
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a601"
	gbp := func(amount string) *moneyrecapv1.Money {
		return &moneyrecapv1.Money{Amount: amount, CurrencyCode: "GBP"}
	}
	ctx := seed(t, db, id, "recaps", "Europe/London", "2026-04-01", "GBP",
		[]*moneyrecapv1.SalarySetting{{Amount: gbp("2400.00"), OriginalAmount: gbp("2400.00"), Payday: "25", EffectiveFrom: "2026-06"}},
		[]*moneyrecapv1.SavingTargetSetting{{Percent: 20, EffectiveFrom: "2026-06"}})
	var rows []*moneyrecapv1.Spend
	for _, s := range []struct {
		amount, date string
		c            moneyrecapv1.Category
	}{
		{"50.00", "2026-04-03", 1},
		{"1500.00", "2026-07-05", 3}, {"500.00", "2026-07-10", 1},
		{"400.00", "2026-08-01", 5}, {"400.00", "2026-08-05", 4}, {"400.00", "2026-08-09", 3}, {"400.00", "2026-08-13", 2}, {"400.00", "2026-08-17", 1},
		{"20.00", "2026-09-03", 1},
	} {
		rows = append(rows, &moneyrecapv1.Spend{Amount: gbp(s.amount), OriginalAmount: gbp(s.amount), Category: s.c, Date: s.date})
	}
	seedSpends(t, db, id, rows)
	clock := at(t, "2026-09-15", "Europe/London")
	s := &Service{DB: db, Now: func() time.Time { return clock }}

	home, err := s.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: "2026-09"}))
	if err != nil {
		t.Fatal(err)
	}
	var list []string
	for _, r := range home.Msg.RecapMonths {
		list = append(list, r.Month+" "+r.Label)
	}
	if want := []string{"2026-08 Slow Burner", "2026-07 Treat-Yourself", "2026-06 Squirrel", "2026-04 Partial"}; !slices.Equal(list, want) {
		t.Errorf("Home's list %v, want %v: May has nothing and September is running", list, want)
	}

	recap := func(month string) (*moneyrecapv1.Recap, error) {
		res, err := s.GetRecap(ctx, connect.NewRequest(&moneyrecapv1.GetRecapRequest{MemberId: id, Month: month}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Recap, nil
	}
	now, err := recap("2026-09")
	if err != nil || now.Situation != moneyrecapv1.PartialSituation_PARTIAL_SITUATION_CURRENT_MONTH || now.ClosingLine != "Your recap arrives on 1 October." ||
		now.Streak.WindowEnd != "2026-09-15" || now.Streak.RunStart != "2026-09-04" || now.Streak.RunEnd != "2026-09-15" || now.Streak.Days != 12 || now.Archetype != nil {
		t.Errorf("the current month: %v, %v, want the preview to today", now, err)
	}
	april, err := recap("2026-04")
	if err != nil || april.Situation != moneyrecapv1.PartialSituation_PARTIAL_SITUATION_NO_MONEY_IN || april.ClosingLine != "Set your monthly salary to get a full recap." || april.Salary != nil {
		t.Errorf("April: %v, %v, want a Partial Recap with no Money In", april, err)
	}
	august, err := recap("2026-08")
	if err != nil || august.Categories[0].Category != moneyrecapv1.Category_CATEGORY_SNACKS || august.Categories[4].Category != moneyrecapv1.Category_CATEGORY_CLOTHES ||
		august.Splurge.Date != "2026-08-01" || august.GetKeptPercent() != 83 {
		t.Errorf("August: %v, %v, want equal shares in D20 order, the earliest Splurge and 83%% kept", august, err)
	}
	for _, month := range []string{"2026-10", "September"} {
		if _, err := recap(month); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", month, err)
		}
	}
	for _, month := range []string{"2026-05", "2026-03"} {
		if _, err := recap(month); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("%s: %v, want not_found for a month without a Recap", month, err)
		}
	}
	withMembers(t, s, nil)
	stranger := auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})
	if _, err := s.GetRecap(stranger, connect.NewRequest(&moneyrecapv1.GetRecapRequest{MemberId: id, Month: "2026-08"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another token: %v, want not_found", err)
	}
}

// kept% is clamped 0 to 999 (PRD §8.4): a 1% target kept a hundred times over, and a month spent past its Salary.
func TestKeptIsClamped(t *testing.T) {
	c, _ := currencyOf("GBP")
	today := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	l := ledgerOf{cash: c, signedUp: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
		salaries: []salaryRow{{amount: big.NewRat(100, 1), payday: "1", from: month{2026, time.August}}},
		targets:  []targetRow{{percent: 1, from: month{2026, time.August}}},
		spends: []spendRow{{amount: big.NewRat(200, 1), date: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC),
			spend: &moneyrecapv1.Spend{Category: moneyrecapv1.Category_CATEGORY_BILLS}}}}
	if got := l.recap(month{2026, time.August}, today).GetKeptPercent(); got != 999 {
		t.Errorf("August kept %d%%, want 999", got)
	}
	if got := l.recap(month{2026, time.September}, today); got.GetKeptPercent() != 0 || got.GetArchetypeName() != "Treat-Yourself" || got.GetArchetypeTitle() != "Treat-Yourself" {
		t.Errorf("September: %v, want 0%% kept and one big treat", got)
	}
}
