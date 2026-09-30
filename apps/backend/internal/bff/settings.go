package bff

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// GetSettings is the Settings view (PRD §6.16): members' Profile, then ledger's settings, each required with its own
// timeout and its code passed through. Notification preferences join later (spec #169); until then there are none.
func (s *Service) GetSettings(ctx context.Context, _ *connect.Request[moneyrecapv1.BffServiceGetSettingsRequest]) (*connect.Response[moneyrecapv1.BffServiceGetSettingsResponse], error) {
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	held, err := s.ledgerSettings(ctx, p)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceGetSettingsResponse{Settings: settingsView(p, held)}), nil
}

// settingsView joins the Profile and ledger's settings, the Currency and its decimals taken from the code ledger holds
// the amounts in, so no view pairs converted amounts with members' older code (spec #169).
func settingsView(p *moneyrecapv1.Profile, ledger *moneyrecapv1.LedgerSettings) *moneyrecapv1.SettingsView {
	profile := proto.CloneOf(p)
	profile.CurrencyCode, profile.CurrencyDigits = heldIn(ledger.GetCurrencyCode(), p)
	return &moneyrecapv1.SettingsView{Profile: profile, LedgerSettings: ledger}
}
