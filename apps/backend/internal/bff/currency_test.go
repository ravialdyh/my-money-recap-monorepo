package bff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
)

// fakeCurrency answers Quote with quote or err, remembering the request.
type fakeCurrency struct {
	moneyrecapv1connect.UnimplementedCurrencyServiceHandler
	quote *moneyrecapv1.QuoteResponse
	err   error
	got   *moneyrecapv1.QuoteRequest
}

func (f *fakeCurrency) Quote(_ context.Context, req *connect.Request[moneyrecapv1.QuoteRequest]) (*connect.Response[moneyrecapv1.QuoteResponse], error) {
	f.got = req.Msg
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(f.quote), nil
}

// currencyClient serves f over HTTP, so codes and details cross the wire as they do in the stack.
func currencyClient(t *testing.T, f *fakeCurrency) moneyrecapv1connect.CurrencyServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewCurrencyServiceHandler(f))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return moneyrecapv1connect.NewCurrencyServiceClient(srv.Client(), srv.URL)
}

// The preview: nothing_to_convert with no quote; ledger's preview of a typed rate as MANUAL, unquoted, or of
// currency's as PROVIDER with its date; the unsupported pair; unavailable for any other failure; no Currency refused.
func TestPreviewCurrencyChange(t *testing.T) {
	provider := &moneyrecapv1.QuoteResponse{Rate: "0.0000496", Source: moneyrecapv1.RateSource_RATE_SOURCE_PROVIDER, AsOf: "2026-09-16"}
	preview := func(amounts bool, to, rate string, quoted *fakeCurrency) (*moneyrecapv1.CurrencyChangePreview, *fakeLedger, error) {
		ledger := &fakeLedger{code: "IDR", amounts: amounts}
		s := newService(t, &fakeMembers{profile: salaryAt(4)}, &fakeNotify{})
		s.Ledger, s.Currency = ledgerClient(t, ledger), currencyClient(t, quoted)
		res, err := s.PreviewCurrencyChange(ctx, connect.NewRequest(&moneyrecapv1.BffServicePreviewCurrencyChangeRequest{ToCurrencyCode: to, ManualRate: rate}))
		if err != nil {
			return nil, ledger, err
		}
		return res.Msg.Preview, ledger, nil
	}
	quoted := &fakeCurrency{quote: provider}
	if p, ledger, err := preview(false, "GBP", "", quoted); err != nil || !p.NothingToConvert || p.FromCurrencyCode != "IDR" || p.ToCurrencyCode != "GBP" || p.Preview != nil || ledger.gotPreview != nil || quoted.got != nil {
		t.Errorf("nothing to convert: %v, %v; want nothing_to_convert from IDR to GBP with no quote and no preview", p, err)
	}
	if p, ledger, err := preview(true, "GBP", "0.0000496", quoted); err != nil || p.NothingToConvert || p.Source != moneyrecapv1.RateSource_RATE_SOURCE_MANUAL ||
		p.AsOf != "" || p.Preview.GetRate() != "0.0000496" || ledger.gotPreview.GetMemberId() != "m-1" || quoted.got != nil {
		t.Errorf("a typed rate: %v, %v; want ledger's preview as MANUAL and no quote", p, err)
	}
	if p, ledger, err := preview(true, "GBP", "", quoted); err != nil || p.Source != moneyrecapv1.RateSource_RATE_SOURCE_PROVIDER || p.AsOf != "2026-09-16" ||
		ledger.gotPreview.GetRate() != "0.0000496" || quoted.got.GetFromCurrencyCode() != "IDR" || quoted.got.GetToCurrencyCode() != "GBP" {
		t.Errorf("a fetched rate: %v, %v, asked %v; want ledger's preview of IDR to GBP as PROVIDER dated 2026-09-16", p, err, quoted.got)
	}
	unsupported := &fakeCurrency{err: connect.NewError(connect.CodeFailedPrecondition, errLedgerDown)}
	if p, ledger, err := preview(true, "GBP", "", unsupported); err != nil || !p.UnsupportedPair || p.Preview != nil || ledger.gotPreview != nil {
		t.Errorf("an unsupported pair: %v, %v; want unsupported_pair with no preview", p, err)
	}
	for _, failed := range []error{connect.NewError(connect.CodeUnavailable, errLedgerDown), connect.NewError(connect.CodeInternal, errLedgerDown)} {
		if _, _, err := preview(true, "GBP", "", &fakeCurrency{err: failed}); connect.CodeOf(err) != connect.CodeUnavailable {
			t.Errorf("currency failing with %v: %v, want unavailable, the failed fetch", failed, err)
		}
	}
	s := newService(t, &fakeMembers{profile: salaryAt(4)}, &fakeNotify{})
	s.Ledger, s.Currency = ledgerClient(t, &fakeLedger{code: "IDR", amounts: true}), moneyrecapv1connect.NewCurrencyServiceClient(http.DefaultClient, "http://127.0.0.1:1")
	if _, err := s.PreviewCurrencyChange(ctx, connect.NewRequest(&moneyrecapv1.BffServicePreviewCurrencyChangeRequest{ToCurrencyCode: "GBP"})); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("currency not answering: %v, want unavailable", err)
	}
	if _, _, err := preview(false, "XXX", "", quoted); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no Currency: %v, want invalid_argument", err)
	}
}

// With a rate ChangeCurrency runs ledger's conversion under the phone's key and answers Home and Settings in the new
// Currency though members still holds the old one; with none it saves members' code, then relabels, under that key.
func TestChangeCurrency(t *testing.T) {
	members, ledger := &fakeMembers{profile: salaryAt(4)}, &fakeLedger{code: "IDR", amounts: true}
	s := newService(t, members, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	msg := &moneyrecapv1.BffServiceChangeCurrencyRequest{ToCurrencyCode: "GBP", Rate: "0.0000496", Source: moneyrecapv1.RateSource_RATE_SOURCE_MANUAL}
	if _, err := s.ChangeCurrency(ctx, keyed(msg, "")); connect.CodeOf(err) != connect.CodeInvalidArgument || ledger.gotConvert != nil {
		t.Fatalf("no key: %v, want invalid_argument before any call", err)
	}
	res, err := s.ChangeCurrency(ctx, keyed(msg, "k1"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ledger.gotConvert; got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.Rate != msg.Rate || got.Source != msg.Source || members.gotProfile != nil {
		t.Errorf("ledger got %v, members %v; want the conversion under k1 and no members write", got, members.gotProfile)
	}
	if v := res.Msg; v.Home.CurrencyCode != "GBP" || v.Home.CurrencyDigits != 2 || v.Home.Spent.CurrencyCode != "GBP" || v.Settings.Profile.CurrencyCode != "GBP" {
		t.Errorf("answer %v, want Home and Settings in GBP at 2 decimals, members' IDR notwithstanding", v)
	}
	if _, err := s.ChangeCurrency(ctx, keyed(&moneyrecapv1.BffServiceChangeCurrencyRequest{ToCurrencyCode: "USD"}, "k2")); connect.CodeOf(err) != connect.CodeFailedPrecondition || members.gotProfile != nil {
		t.Errorf("no rate with amounts: %v, members %v; want failed_precondition and no write", err, members.gotProfile)
	}

	members, ledger = &fakeMembers{profile: salaryAt(4)}, &fakeLedger{code: "IDR"}
	s = newService(t, members, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	res, err = s.ChangeCurrency(ctx, keyed(&moneyrecapv1.BffServiceChangeCurrencyRequest{ToCurrencyCode: "USD"}, "k3"))
	if err != nil {
		t.Fatal(err)
	}
	if m, l := members.gotProfile, ledger.gotRelabel; m.GetIdempotencyKey() != "k3" || m.GetCurrencyCode() != "USD" || m.OnboardingStep != nil ||
		l.GetIdempotencyKey() != "k3" || l.GetCurrencyCode() != "USD" || ledger.gotConvert != nil {
		t.Errorf("members got %v, ledger %v; want USD saved, then relabelled, under k3", m, l)
	}
	if v := res.Msg; v.Settings.Profile.CurrencyCode != "USD" || v.Settings.LedgerSettings.CurrencyCode != "USD" {
		t.Errorf("answer %v, want Settings in USD", v)
	}
	if _, err := s.ChangeCurrency(ctx, keyed(&moneyrecapv1.BffServiceChangeCurrencyRequest{ToCurrencyCode: "XXX"}, "k4")); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no Currency: %v, want invalid_argument", err)
	}
	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeOutOfRange, errLedgerDown)})
	if _, err := s.ChangeCurrency(ctx, keyed(msg, "k5")); connect.CodeOf(err) != connect.CodeOutOfRange {
		t.Errorf("ledger refusing past D57: %v, want out_of_range", err)
	}
}
