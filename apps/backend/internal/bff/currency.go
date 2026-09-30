package bff

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/answers"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/cldr"
)

// errNoQuote is PreviewCurrencyChange without a typed rate when currency does not quote one: the phone shows the
// failed fetch and "Enter it myself" (PRD §8.9 step 1).
var errNoQuote = errors.New("no rate could be fetched: enter the rate")

// errNeedsRate refuses the direct change for a Member with amounts, which only a conversion may move (PRD §8.9).
var errNeedsRate = errors.New("there are amounts to convert: send the rate the confirm sheet showed")

// PreviewCurrencyChange is the Currency sheet's preview (PRD §8.9 steps 1 to 3): nothing_to_convert with no quote for
// a Member with no amounts, else ledger's preview of the typed rate as MANUAL or of currency.Quote's as PROVIDER.
func (s *Service) PreviewCurrencyChange(ctx context.Context, req *connect.Request[moneyrecapv1.BffServicePreviewCurrencyChangeRequest]) (*connect.Response[moneyrecapv1.BffServicePreviewCurrencyChangeResponse], error) {
	to := req.Msg.ToCurrencyCode
	if err := answers.CurrencyCode(to); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	held, err := s.ledgerSettings(ctx, p)
	if err != nil {
		return nil, err
	}
	preview := &moneyrecapv1.CurrencyChangePreview{FromCurrencyCode: held.CurrencyCode, ToCurrencyCode: to, NothingToConvert: !held.HasAmountsToConvert}
	if preview.NothingToConvert {
		return connect.NewResponse(&moneyrecapv1.BffServicePreviewCurrencyChangeResponse{Preview: preview}), nil
	}
	rate := req.Msg.ManualRate
	preview.Source = moneyrecapv1.RateSource_RATE_SOURCE_MANUAL
	if rate == "" {
		quote, err := s.quote(ctx, held.CurrencyCode, to)
		if connect.CodeOf(err) == connect.CodeFailedPrecondition {
			preview.UnsupportedPair = true
			return connect.NewResponse(&moneyrecapv1.BffServicePreviewCurrencyChangeResponse{Preview: preview}), nil
		}
		if err != nil {
			return nil, err
		}
		rate, preview.Source, preview.AsOf = quote.Rate, quote.Source, quote.AsOf
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	res, err := s.Ledger.PreviewConversion(ctx, connect.NewRequest(&moneyrecapv1.PreviewConversionRequest{MemberId: p.MemberId, Rate: rate, ToCurrencyCode: to}))
	if err != nil {
		return nil, requiredCallFailed("ledger.PreviewConversion", connect.CodeOf(err), err)
	}
	preview.Preview = res.Msg.Preview
	return connect.NewResponse(&moneyrecapv1.BffServicePreviewCurrencyChangeResponse{Preview: preview}), nil
}

// quote is currency.Quote, optional: its failed_precondition, the unsupported pair, passes through, and any other
// failure or no answer is unavailable, the failed fetch.
func (s *Service) quote(ctx context.Context, from, to string) (*moneyrecapv1.QuoteResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	res, err := s.Currency.Quote(ctx, connect.NewRequest(&moneyrecapv1.QuoteRequest{FromCurrencyCode: from, ToCurrencyCode: to}))
	switch {
	case connect.CodeOf(err) == connect.CodeFailedPrecondition:
		return nil, err
	case err != nil:
		logOptionalCallDegraded("currency.Quote", "the failed fetch", err)
		return nil, connect.NewError(connect.CodeUnavailable, errNoQuote)
	}
	return res.Msg, nil
}

// ChangeCurrency is the confirm sheet's Convert, ledger's conversion under the phone's key, or with no rate the direct
// change: members' Currency, then ledger's relabel, under the one key. It answers a fresh Home and Settings.
func (s *Service) ChangeCurrency(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceChangeCurrencyRequest]) (*connect.Response[moneyrecapv1.BffServiceChangeCurrencyResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	to := req.Msg.ToCurrencyCode
	if err := answers.CurrencyCode(to); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	var settings *moneyrecapv1.LedgerSettings
	if req.Msg.Rate == "" {
		p, settings, err = s.changeDirectly(ctx, key, p, to)
	} else {
		settings, err = s.convert(ctx, key, p, req.Msg)
	}
	if err != nil {
		return nil, err
	}
	home, err := s.home(ctx, p, nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceChangeCurrencyResponse{Home: home, Settings: settingsView(p, settings)}), nil
}

// convert is ledger.ApplyCurrencyConversion under key, its code passed through: out_of_range past D57.
func (s *Service) convert(ctx context.Context, key string, p *moneyrecapv1.Profile, msg *moneyrecapv1.BffServiceChangeCurrencyRequest) (*moneyrecapv1.LedgerSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	res, err := s.Ledger.ApplyCurrencyConversion(ctx, connect.NewRequest(&moneyrecapv1.ApplyCurrencyConversionRequest{
		IdempotencyKey: key, MemberId: p.MemberId, Rate: msg.Rate, Source: msg.Source, ToCurrencyCode: msg.ToCurrencyCode,
	}))
	if err != nil {
		return nil, requiredCallFailed("ledger.ApplyCurrencyConversion", connect.CodeOf(err), err)
	}
	return res.Msg.Settings, nil
}

// changeDirectly is PRD §8.9 step 2, setup's two writes: members.SaveProfile, then ledger's relabel, so both hold the
// new code before the answer. failed_precondition while ledger holds amounts.
func (s *Service) changeDirectly(ctx context.Context, key string, p *moneyrecapv1.Profile, to string) (*moneyrecapv1.Profile, *moneyrecapv1.LedgerSettings, error) {
	held, err := s.ledgerSettings(ctx, p)
	if err != nil {
		return nil, nil, err
	}
	if held.HasAmountsToConvert {
		return nil, nil, connect.NewError(connect.CodeFailedPrecondition, errNeedsRate)
	}
	membersCtx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	saved, err := s.Members.SaveProfile(membersCtx, connect.NewRequest(&moneyrecapv1.SaveProfileRequest{IdempotencyKey: key, MemberId: p.MemberId, CurrencyCode: proto.String(to)}))
	if err != nil {
		return nil, nil, requiredCallFailed("members.SaveProfile", connect.CodeOf(err), err)
	}
	settings, err := s.relabel(ctx, key, p, to)
	if err != nil {
		return nil, nil, err
	}
	return saved.Msg.Profile, settings, nil
}

// ledgerSettings is ledger.GetSettings for the Member, required, its code passed through.
func (s *Service) ledgerSettings(ctx context.Context, p *moneyrecapv1.Profile) (*moneyrecapv1.LedgerSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	res, err := s.Ledger.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: p.MemberId}))
	if err != nil {
		return nil, requiredCallFailed("ledger.GetSettings", connect.CodeOf(err), err)
	}
	return res.Msg.Settings, nil
}

// heldIn is a view's Currency and its decimals (spec #169): code, the one ledger answered the amounts in, or the
// Profile's when ledger named none, so no view pairs converted amounts with members' older code.
func heldIn(code string, p *moneyrecapv1.Profile) (held string, digits int32) {
	if n, ok := cldr.Digits(code); ok {
		return code, int32(n)
	}
	return p.GetCurrencyCode(), p.GetCurrencyDigits()
}
