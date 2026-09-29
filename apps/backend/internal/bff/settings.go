package bff

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/cldr"
)

// GetSettings is the Settings view (PRD §6.16): members' Profile, then ledger's settings, each required with its own
// timeout and its code passed through. Notification preferences join later (spec #169); until then there are none.
func (s *Service) GetSettings(ctx context.Context, _ *connect.Request[moneyrecapv1.BffServiceGetSettingsRequest]) (*connect.Response[moneyrecapv1.BffServiceGetSettingsResponse], error) {
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	res, err := s.Ledger.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: p.MemberId}))
	if err != nil {
		return nil, requiredCallFailed("ledger.GetSettings", connect.CodeOf(err), err)
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceGetSettingsResponse{Settings: settingsView(p, res.Msg.Settings)}), nil
}

// settingsView joins the Profile and ledger's settings, the Currency and its decimals taken from the code ledger holds
// the amounts in, so no view pairs converted amounts with members' older code (spec #169).
func settingsView(p *moneyrecapv1.Profile, ledger *moneyrecapv1.LedgerSettings) *moneyrecapv1.SettingsView {
	profile := proto.CloneOf(p)
	if digits, ok := cldr.Digits(ledger.GetCurrencyCode()); ok {
		profile.CurrencyCode, profile.CurrencyDigits = ledger.GetCurrencyCode(), int32(digits)
	}
	return &moneyrecapv1.SettingsView{Profile: profile, LedgerSettings: ledger}
}
