package bff

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// GetSettings joins the Profile and ledger's settings, the Currency and its decimals ledger's; either service failing
// fails the read with its own code and no upstream text (spec #169).
func TestGetSettingsJoinsProfileAndLedger(t *testing.T) {
	target := &moneyrecapv1.SavingTargetSetting{Percent: 20, EffectiveFrom: "2026-09"}
	ledger := &fakeLedger{target: target, code: "BHD"}
	s := newService(t, &fakeMembers{profile: salaryAt(4)}, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	res, err := s.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetSettingsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if v := res.Msg.Settings; v.Profile.DisplayName != "Maya" || v.Profile.CurrencyCode != "BHD" || v.Profile.CurrencyDigits != 3 ||
		v.LedgerSettings.SavingTarget.GetPercent() != 20 || v.LedgerSettings.Salary != nil || v.NotificationPrefs != nil {
		t.Errorf("Settings %v, want Maya in ledger's BHD at 3 decimals, 20%%, no Salary and no preferences", v)
	}

	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeNotFound, errLedgerDown)})
	if _, err := s.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetSettingsRequest{})); connect.CodeOf(err) != connect.CodeNotFound || strings.Contains(err.Error(), "172.18") {
		t.Errorf("ledger refusing: %v, want not_found naming no host", err)
	}
	down := newService(t, &fakeMembers{createErr: connect.NewError(connect.CodeUnavailable, errLedgerDown)}, &fakeNotify{})
	if _, err := down.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetSettingsRequest{})); connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "172.18") {
		t.Errorf("members down: %v, want unavailable naming no host", err)
	}
}

// A Salary or Target save after setup answers the Home GetHome draws: the past-Recap list, notify's kept banner and
// the earliest Spend date, the month read in the Member's zone, 21:00 UTC on 30 September being October in Jakarta.
func TestSalaryAndTargetSavesAnswerTheWholeHome(t *testing.T) {
	p := salaryAt(4)
	p.Timezone = "Asia/Jakarta"
	kept := &moneyrecapv1.TargetWarning{Message: "You've passed your saving line. Rp50.000 over."}
	recaps := []*moneyrecapv1.RecapMonth{{Month: "2026-09", Label: "The Hustler"}}
	ledger := &fakeLedger{recaps: recaps, code: "IDR"}
	s := newService(t, &fakeMembers{profile: p}, &fakeNotify{kept: kept})
	s.Ledger = ledgerClient(t, ledger)
	s.Now = func() time.Time { return time.Date(2026, 9, 30, 21, 0, 0, 0, time.UTC) }
	whole := func(what string, h *moneyrecapv1.HomeView) {
		t.Helper()
		if h.Month != "2026-10" || ledger.gotMonth != "2026-10" || len(h.RecapMonths) != 1 || h.TargetWarning.GetMessage() != kept.Message || h.EarliestSpendDate != "2026-07-19" {
			t.Errorf("%s answered Home %v after asking ledger for %q, want October with the Recap list, the banner and the earliest date", what, h, ledger.gotMonth)
		}
	}
	salary, err := s.SaveSalary(ctx, keyed(&moneyrecapv1.BffServiceSaveSalaryRequest{Amount: &moneyrecapv1.Money{Amount: "15000000", CurrencyCode: "IDR"}, Payday: "25"}, "k1"))
	if err != nil {
		t.Fatal(err)
	}
	whole("SaveSalary", salary.Msg.Home)
	target, err := s.SaveTarget(ctx, keyed(&moneyrecapv1.BffServiceSaveTargetRequest{Percent: 20}, "k2"))
	if err != nil {
		t.Fatal(err)
	}
	whole("SaveTarget", target.Msg.Home)
}
