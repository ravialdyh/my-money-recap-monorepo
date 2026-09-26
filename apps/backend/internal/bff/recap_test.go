package bff

import (
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// GetRecap joins ledger's month with the Member's current Display Name and Currency under the caller's token, and
// passes ledger's refusals through naming only the call (PRD §7, ARCHITECTURE §6).
func TestGetRecapJoinsTheMemberAndPassesCodesThrough(t *testing.T) {
	ledger := &fakeLedger{}
	s := newService(t, &fakeMembers{profile: onboarded()}, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	get := func() (*connect.Response[moneyrecapv1.BffServiceGetRecapResponse], error) {
		return s.GetRecap(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetRecapRequest{Month: "2026-08"}))
	}
	res, err := get()
	if err != nil {
		t.Fatal(err)
	}
	if v := res.Msg.Recap; v.Recap.GetMonth() != "2026-08" || v.Recap.GetArchetypeTitle() != "The Squirrel" || v.DisplayName != "Maya" || v.CurrencyCode != "IDR" {
		t.Errorf("RecapView %v, want ledger's August as Maya in IDR", v)
	}
	if ledger.gotMonth != "2026-08" || ledger.gotToken != "Bearer tok" {
		t.Errorf("ledger got %q with %q, want 2026-08 with the caller's token", ledger.gotMonth, ledger.gotToken)
	}
	for _, code := range []connect.Code{connect.CodeNotFound, connect.CodeInvalidArgument, connect.CodeUnavailable} {
		s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(code, errLedgerDown)})
		if _, err := get(); connect.CodeOf(err) != code || err.Error() != code.String()+": ledger.GetRecap failed" {
			t.Errorf("ledger's %v: %v, want it passed through naming only the call", code, err)
		}
	}
	s = newService(t, &fakeMembers{createErr: connect.NewError(connect.CodeUnavailable, errLedgerDown)}, &fakeNotify{})
	if _, err := get(); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("members down: %v, want unavailable", err)
	}
}
