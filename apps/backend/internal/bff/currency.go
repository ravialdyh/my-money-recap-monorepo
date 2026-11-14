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

var errNoQuote = errors.New("no rate could be fetched: enter the rate")

var errNeedsRate = errors.New("there are amounts to convert: send the rate the confirm sheet showed")

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

func (s *Service) ledgerSettings(ctx context.Context, p *moneyrecapv1.Profile) (*moneyrecapv1.LedgerSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	res, err := s.Ledger.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: p.MemberId}))
	if err != nil {
		return nil, requiredCallFailed("ledger.GetSettings", connect.CodeOf(err), err)
	}
	return res.Msg.Settings, nil
}

func heldIn(code string, p *moneyrecapv1.Profile) (held string, digits int32) {
	if n, ok := cldr.Digits(code); ok {
		return code, int32(n)
	}
	return p.GetCurrencyCode(), p.GetCurrencyDigits()
}
