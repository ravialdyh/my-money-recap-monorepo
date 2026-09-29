package bff

import (
	"context"
	"sync"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// GetSettings is the Settings view (PRD §6.16): members' Profile, then ledger's settings, required, and notify's
// preferences, optional, asked at once with their own timeouts. Without notify the view carries no preferences.
func (s *Service) GetSettings(ctx context.Context, _ *connect.Request[moneyrecapv1.BffServiceGetSettingsRequest]) (*connect.Response[moneyrecapv1.BffServiceGetSettingsResponse], error) {
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	var (
		wg                sync.WaitGroup
		held              *moneyrecapv1.LedgerSettings
		prefs             *connect.Response[moneyrecapv1.GetPrefsResponse]
		heldErr, prefsErr error
	)
	wg.Go(func() { held, heldErr = s.ledgerSettings(ctx, p) })
	wg.Go(func() {
		ctx, cancel := context.WithTimeout(ctx, s.timeout())
		defer cancel()
		prefs, prefsErr = s.Notify.GetPrefs(ctx, connect.NewRequest(&moneyrecapv1.GetPrefsRequest{MemberId: p.MemberId}))
	})
	wg.Wait()
	if heldErr != nil {
		return nil, heldErr
	}
	view := settingsView(p, held)
	if prefsErr != nil {
		logOptionalCallDegraded("notify.GetPrefs", "SettingsView returned without preferences", prefsErr)
	} else {
		view.NotificationPrefs = prefs.Msg.Prefs
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceGetSettingsResponse{Settings: view}), nil
}

// settingsView joins the Profile and ledger's settings, the Currency and its decimals taken from the code ledger holds
// the amounts in, so no view pairs converted amounts with members' older code (spec #169).
func settingsView(p *moneyrecapv1.Profile, ledger *moneyrecapv1.LedgerSettings) *moneyrecapv1.SettingsView {
	profile := proto.CloneOf(p)
	profile.CurrencyCode, profile.CurrencyDigits = heldIn(ledger.GetCurrencyCode(), p)
	return &moneyrecapv1.SettingsView{Profile: profile, LedgerSettings: ledger}
}

// SaveNotificationPrefs hands the whole preferences message to notify under the phone's key (PRD §6.16, A14) and
// answers it as notify stored it, notify's code passed through: the phone puts the switch back on any failure.
func (s *Service) SaveNotificationPrefs(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceSaveNotificationPrefsRequest]) (*connect.Response[moneyrecapv1.BffServiceSaveNotificationPrefsResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	saved, err := s.Notify.SavePrefs(ctx, connect.NewRequest(&moneyrecapv1.SavePrefsRequest{IdempotencyKey: key, MemberId: profile.MemberId, Prefs: req.Msg.Prefs}))
	if err != nil {
		return nil, requiredCallFailed("notify.SavePrefs", connect.CodeOf(err), err)
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceSaveNotificationPrefsResponse{Prefs: saved.Msg.Prefs}), nil
}
