package bff

import (
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// GetSettings joins the Profile, ledger's settings with its Currency and decimals, and notify's preferences; a required
// service failing fails the read with its code and no upstream text, notify silent leaves out the preferences.
func TestGetSettingsJoinsProfileLedgerAndPrefs(t *testing.T) {
	target := &moneyrecapv1.SavingTargetSetting{Percent: 20, EffectiveFrom: "2026-09"}
	ledger := &fakeLedger{target: target, code: "BHD"}
	prefs := &moneyrecapv1.NotificationPrefs{RecapReady: true, ReminderTime: "20:00", TargetWarning: true}
	notify := &fakeNotify{prefs: prefs}
	s := newService(t, &fakeMembers{profile: salaryAt(4)}, notify)
	s.Ledger = ledgerClient(t, ledger)
	res, err := s.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetSettingsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if v := res.Msg.Settings; v.Profile.DisplayName != "Maya" || v.Profile.CurrencyCode != "BHD" || v.Profile.CurrencyDigits != 3 ||
		v.LedgerSettings.SavingTarget.GetPercent() != 20 || v.LedgerSettings.Salary != nil || !proto.Equal(v.NotificationPrefs, prefs) {
		t.Errorf("Settings %v, want Maya in ledger's BHD at 3 decimals, 20%%, no Salary and notify's preferences", v)
	}
	notify.hang = true
	res, err = s.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetSettingsRequest{}))
	if err != nil || res.Msg.Settings.NotificationPrefs != nil || res.Msg.Settings.LedgerSettings.SavingTarget.GetPercent() != 20 {
		t.Errorf("notify not answering: %v (%v), want the view without preferences", res, err)
	}
	notify.hang = false

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
	salary, err := s.SaveSalary(ctx, keyed(&moneyrecapv1.BffServiceSaveSalaryRequest{Amount: &moneyrecapv1.Money{Amount: "150000000", CurrencyCode: "IDR"}, Payday: "25"}, "k1"))
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

// SaveNotificationPrefs refuses a keyless call before notify, then hands notify the key, the Member's id and the whole
// message, answering what notify stored; notify's refusal comes back with its code and no upstream text.
func TestSaveNotificationPrefsPassesTheWholeMessageThrough(t *testing.T) {
	notify := &fakeNotify{}
	s := newService(t, &fakeMembers{profile: salaryAt(4)}, notify)
	prefs := &moneyrecapv1.NotificationPrefs{RecapReady: true, DailyReminder: true, ReminderTime: "07:30"}
	req := &moneyrecapv1.BffServiceSaveNotificationPrefsRequest{Prefs: prefs}
	if _, err := s.SaveNotificationPrefs(ctx, keyed(req, "")); connect.CodeOf(err) != connect.CodeInvalidArgument || notify.saved != nil {
		t.Fatalf("no key: %v, notify asked %v; want invalid_argument before notify", err, notify.saved)
	}
	res, err := s.SaveNotificationPrefs(ctx, keyed(req, "key-1"))
	if err != nil || !proto.Equal(res.Msg.Prefs, prefs) || notify.saved.IdempotencyKey != "key-1" || notify.saved.MemberId != "m-1" || !proto.Equal(notify.saved.Prefs, prefs) {
		t.Fatalf("SaveNotificationPrefs: %v (%v); notify got %v", res, err, notify.saved)
	}
	notify.saveErr = connect.NewError(connect.CodeUnavailable, errLedgerDown)
	if _, err := s.SaveNotificationPrefs(ctx, keyed(req, "key-2")); connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "172.18") {
		t.Errorf("notify refusing: %v, want unavailable naming no host", err)
	}
}
