package bff

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// LogSpend hands the Spend to ledger under the phone's key and the caller's token, then answers the current month's
// Home in the Member's Currency and decimals and ledger's earliest date, even for a Spend dated last month; ledger's refusal keeps its code.
func TestLogSpendWritesLedgerThenAnswersThisMonthsHome(t *testing.T) {
	msg := &moneyrecapv1.BffServiceLogSpendRequest{
		Amount: &moneyrecapv1.Money{Amount: "12.40", CurrencyCode: "GBP"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: "2026-08-30", Note: "Kopi",
	}
	profile := onboarded()
	profile.CurrencyCode, profile.CurrencyDigits = "GBP", 2
	ledger := &fakeLedger{}
	s := newService(t, &fakeMembers{profile: profile}, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	if _, err := s.LogSpend(ctx, keyed(msg, "")); connect.CodeOf(err) != connect.CodeInvalidArgument || ledger.gotSpend != nil {
		t.Fatalf("no key: %v, ledger reached %v; want invalid_argument before any call", err, ledger.gotSpend != nil)
	}
	res, err := s.LogSpend(ctx, keyed(msg, "k1"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ledger.gotSpend; got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.Date != "2026-08-30" || got.Note != "Kopi" ||
		got.Category != msg.Category || got.Amount.Amount != "12.40" || ledger.gotToken != "Bearer tok" {
		t.Errorf("ledger got %v with %q, want the Spend under k1 for m-1 with the caller's token", got, ledger.gotToken)
	}
	if h := res.Msg.Home; h.Month != "2026-09" || ledger.gotMonth != "2026-09" || h.CurrencyCode != "GBP" || h.CurrencyDigits != 2 || h.Spent.CurrencyCode != "GBP" || h.EarliestSpendDate != "2026-07-19" {
		t.Errorf("Home %v read for %q, want September's in GBP with 2 decimals and ledger's earliest date", h, ledger.gotMonth)
	}
	if res.Msg.Spend.GetId() != "s-1" {
		t.Errorf("Spend %v, want ledger's", res.Msg.Spend)
	}

	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeInvalidArgument, errLedgerDown)})
	if _, err := s.LogSpend(ctx, keyed(msg, "k2")); connect.CodeOf(err) != connect.CodeInvalidArgument || err.Error() != "invalid_argument: ledger.LogSpend failed" {
		t.Errorf("a refused Spend: %v, want invalid_argument naming only the call", err)
	}
}

// A current-month Spend tells notify ledger's crossed_line and copy under the phone's key, and Home's banner is the
// copy notify kept; a past-month Spend stamps nothing and reads the state; notify hanging still answers, without it.
func TestLogSpendEvaluatesTheWarningForThisMonthOnly(t *testing.T) {
	first, later := &moneyrecapv1.TargetWarning{Message: "You've passed your saving line. Rp50.000 over."}, &moneyrecapv1.TargetWarning{Message: "You've passed your saving line. Rp150.000 over."}
	notify := &fakeNotify{}
	s := newService(t, &fakeMembers{profile: onboarded()}, notify)
	ledger := &fakeLedger{crossed: true, warning: first}
	s.Ledger = ledgerClient(t, ledger)
	spend := func(key, date string) *moneyrecapv1.HomeView {
		t.Helper()
		res, err := s.LogSpend(ctx, keyed(&moneyrecapv1.BffServiceLogSpendRequest{
			Amount: &moneyrecapv1.Money{Amount: "50000", CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: date,
		}, key))
		if err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		return res.Msg.Home
	}

	if h := spend("k1", "2026-09-16"); h.TargetWarning.GetMessage() != first.Message || notify.reads != 0 {
		t.Fatalf("the crossing Spend: banner %v after %d state reads, want the kept copy from EvaluateWarning alone", h.TargetWarning, notify.reads)
	}
	if got := notify.evaluated; got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.Month != "2026-09" || !got.CrossedLine || got.Warning.GetMessage() != first.Message {
		t.Errorf("notify got %v, want k1 for m-1 in 2026-09 with ledger's verdict and copy", got)
	}
	ledger.crossed, ledger.warning = false, later
	if h := spend("k2", "2026-09-16"); h.TargetWarning.GetMessage() != first.Message {
		t.Errorf("a later Spend: banner %v, want the first words kept", h.TargetWarning)
	}
	notify.evaluated = nil
	if h := spend("k3", "2026-08-30"); notify.evaluated != nil || notify.reads != 1 || h.TargetWarning.GetMessage() != first.Message {
		t.Errorf("a Spend in August: evaluated %v, %d state reads, banner %v; want no stamp and the state read", notify.evaluated, notify.reads, h.TargetWarning)
	}
	s.Notify = &fakeNotify{hang: true}
	start := time.Now()
	if h := spend("k4", "2026-09-16"); h.TargetWarning != nil || h.Spent == nil || time.Since(start) > time.Second {
		t.Errorf("notify hanging: %v after %v, want Home without a banner within the timeouts", h, time.Since(start))
	}
}

// ListSpends asks ledger for the month given, or with none the current month in the Member's zone, and joins its Spends,
// the switcher's months, the month's Streak and the Currency's decimals; ledger's refusal keeps its code.
func TestListSpendsJoinsLedgersMonth(t *testing.T) {
	profile := onboarded()
	profile.Timezone, profile.CurrencyDigits = "Asia/Jakarta", 0
	ledger := &fakeLedger{}
	s := newService(t, &fakeMembers{profile: profile}, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	s.Now = func() time.Time { return time.Date(2026, 9, 30, 20, 0, 0, 0, time.UTC) }
	res, err := s.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.BffServiceListSpendsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	v := res.Msg.Spends
	if ledger.gotList != "2026-10" || ledger.gotMonth != "2026-10" || v.Month != "2026-10" || ledger.gotToken != "Bearer tok" {
		t.Errorf("asked %q and %q for %q with %q, want Jakarta's October with the caller's token", ledger.gotList, ledger.gotMonth, v.Month, ledger.gotToken)
	}
	if v.CurrencyCode != "IDR" || v.CurrencyDigits != 0 || v.StreakDay != 1 || len(v.Months) != 2 || v.Spent.Amount != "25000" ||
		len(v.Spends) != 1 || v.Categories[0].SpendCount != 1 {
		t.Errorf("view %v, want ledger's month in IDR with 0 decimals and streak day 1", v)
	}
	if _, err := s.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.BffServiceListSpendsRequest{Month: "2026-08"})); err != nil || ledger.gotList != "2026-08" {
		t.Errorf("August: %v, asked %q, want the month given", err, ledger.gotList)
	}

	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeInvalidArgument, errLedgerDown)})
	_, err = s.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.BffServiceListSpendsRequest{Month: "August"}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument || strings.Contains(err.Error(), "5432") {
		t.Errorf("a refused month: %v, want invalid_argument naming only the call", err)
	}
}
