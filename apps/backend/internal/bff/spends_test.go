package bff

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

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

func TestEditSpendWritesLedgerThenAnswersThisMonthsHome(t *testing.T) {
	kept := &moneyrecapv1.TargetWarning{Message: "You've passed your saving line. Rp50.000 over."}
	notify := &fakeNotify{}
	ledger := &fakeLedger{crossed: true, warning: kept}
	s := newService(t, &fakeMembers{profile: onboarded()}, notify)
	s.Ledger = ledgerClient(t, ledger)
	edit := func(key, date string) (*connect.Response[moneyrecapv1.BffServiceEditSpendResponse], error) {
		return s.EditSpend(ctx, keyed(&moneyrecapv1.BffServiceEditSpendRequest{
			SpendId: "s-1", Amount: &moneyrecapv1.Money{Amount: "50000", CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_BILLS, Date: date, Note: "Listrik",
		}, key))
	}
	if _, err := edit("", "2026-09-16"); connect.CodeOf(err) != connect.CodeInvalidArgument || ledger.gotEdit != nil {
		t.Fatalf("no key: %v, ledger reached %v; want invalid_argument before any call", err, ledger.gotEdit != nil)
	}
	res, err := edit("k1", "2026-09-16")
	if err != nil {
		t.Fatal(err)
	}
	if got := ledger.gotEdit; got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.SpendId != "s-1" || got.Amount.Amount != "50000" ||
		got.Category != moneyrecapv1.Category_CATEGORY_BILLS || got.Date != "2026-09-16" || got.Note != "Listrik" || ledger.gotToken != "Bearer tok" {
		t.Errorf("ledger got %v with %q, want every field under k1 for m-1 with the caller's token", got, ledger.gotToken)
	}
	if got := notify.evaluated; got.GetIdempotencyKey() != "k1" || got.Month != "2026-09" || !got.CrossedLine || got.Warning.GetMessage() != kept.Message {
		t.Errorf("notify got %v, want k1 in 2026-09 with ledger's verdict and copy", got)
	}
	if h := res.Msg.Home; h.Month != "2026-09" || h.TargetWarning.GetMessage() != kept.Message || notify.reads != 0 || res.Msg.Spend.GetId() != "s-1" {
		t.Errorf("answer %v after %d state reads, want September's Home under the kept copy and ledger's Spend", res.Msg, notify.reads)
	}
	notify.evaluated = nil
	if res, err := edit("k2", "2026-08-30"); err != nil || notify.evaluated != nil || notify.reads != 1 || res.Msg.Home.Month != "2026-09" {
		t.Errorf("an edit into August: %v, %v, evaluated %v; want September's Home, no stamp and the state read", res, err, notify.evaluated)
	}
	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeNotFound, errLedgerDown)})
	if _, err := edit("k3", "2026-09-16"); connect.CodeOf(err) != connect.CodeNotFound || err.Error() != "not_found: ledger.EditSpend failed" {
		t.Errorf("another Member's Spend: %v, want not_found naming only the call", err)
	}
}

func TestDeleteSpendKeepsTheBannerFromTheState(t *testing.T) {
	kept := &moneyrecapv1.TargetWarning{Message: "You've passed your saving line. Rp50.000 over."}
	notify := &fakeNotify{kept: kept}
	ledger := &fakeLedger{}
	s := newService(t, &fakeMembers{profile: onboarded()}, notify)
	s.Ledger = ledgerClient(t, ledger)
	del := func(key string) (*connect.Response[moneyrecapv1.BffServiceDeleteSpendResponse], error) {
		return s.DeleteSpend(ctx, keyed(&moneyrecapv1.BffServiceDeleteSpendRequest{SpendId: "s-1"}, key))
	}
	if _, err := del(""); connect.CodeOf(err) != connect.CodeInvalidArgument || ledger.gotDelete != nil {
		t.Fatalf("no key: %v, ledger reached %v; want invalid_argument before any call", err, ledger.gotDelete != nil)
	}
	res, err := del("k1")
	if err != nil {
		t.Fatal(err)
	}
	if got := ledger.gotDelete; got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.SpendId != "s-1" || ledger.gotToken != "Bearer tok" {
		t.Errorf("ledger got %v with %q, want s-1 under k1 for m-1 with the caller's token", got, ledger.gotToken)
	}
	if h := res.Msg.Home; h.Month != "2026-09" || h.TargetWarning.GetMessage() != kept.Message || notify.evaluated != nil || notify.reads != 1 {
		t.Errorf("Home %v, evaluated %v, %d state reads; want September under the kept copy, read and never evaluated", h, notify.evaluated, notify.reads)
	}
	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeNotFound, errLedgerDown)})
	if _, err := del("k2"); connect.CodeOf(err) != connect.CodeNotFound || err.Error() != "not_found: ledger.DeleteSpend failed" {
		t.Errorf("another Member's Spend: %v, want not_found naming only the call", err)
	}
}
