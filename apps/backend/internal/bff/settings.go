package bff

import (
	"context"
	"sync"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

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

func settingsView(p *moneyrecapv1.Profile, ledger *moneyrecapv1.LedgerSettings) *moneyrecapv1.SettingsView {
	profile := proto.CloneOf(p)
	profile.CurrencyCode, profile.CurrencyDigits = heldIn(ledger.GetCurrencyCode(), p)
	return &moneyrecapv1.SettingsView{Profile: profile, LedgerSettings: ledger}
}

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
