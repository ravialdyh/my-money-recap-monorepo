package bff

import (
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// LogSpend hands the Spend to ledger under the phone's key and the caller's token, then answers the current month's
// Home in the Member's Currency and decimals, even for a Spend dated last month; ledger's refusal keeps its code.
func TestLogSpendWritesLedgerThenAnswersThisMonthsHome(t *testing.T) {
	msg := &moneyrecapv1.BffServiceLogSpendRequest{
		Amount: &moneyrecapv1.Money{Amount: "12.40", CurrencyCode: "GBP"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: "2026-08-30", Note: "Kopi",
	}
	profile := onboarded()
	profile.CurrencyCode, profile.CurrencyDigits = "GBP", 2
	ledger := &fakeLedger{}
	s := newService(t, &fakeMembers{profile: profile}, StubNotify{})
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
	if h := res.Msg.Home; h.Month != "2026-09" || ledger.gotMonth != "2026-09" || h.CurrencyCode != "GBP" || h.CurrencyDigits != 2 || h.Spent.CurrencyCode != "GBP" {
		t.Errorf("Home %v read for %q, want September's in GBP with 2 decimals", h, ledger.gotMonth)
	}
	if res.Msg.Spend.GetId() != "s-1" {
		t.Errorf("Spend %v, want ledger's", res.Msg.Spend)
	}

	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeInvalidArgument, errLedgerDown)})
	if _, err := s.LogSpend(ctx, keyed(msg, "k2")); connect.CodeOf(err) != connect.CodeInvalidArgument || err.Error() != "invalid_argument: ledger.LogSpend failed" {
		t.Errorf("a refused Spend: %v, want invalid_argument naming only the call", err)
	}
}
