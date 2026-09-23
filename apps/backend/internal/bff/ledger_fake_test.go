package bff

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// errLedgerDown is how a ledger whose database is gone fails, naming the address.
var errLedgerDown = errors.New("dial tcp 172.18.0.5:5432: connect: connection refused")

// fakeLedger is ledger as the BFF sees it: a zero month, settings holding salary and target, and each write recorded
// with the token it came with; err fails every call.
type fakeLedger struct {
	moneyrecapv1connect.UnimplementedLedgerServiceHandler
	err        error
	salary     *moneyrecapv1.SalarySetting
	target     *moneyrecapv1.SavingTargetSetting
	gotToken   string
	gotSalary  *moneyrecapv1.SaveSalaryRequest
	gotTarget  *moneyrecapv1.SaveTargetRequest
	gotRelabel *moneyrecapv1.RelabelCurrencyRequest
	settings   int
}

func zeroMonth() *moneyrecapv1.MonthSummary {
	return &moneyrecapv1.MonthSummary{Spent: &moneyrecapv1.Money{Amount: "0"}, MoneyIn: &moneyrecapv1.Money{Amount: "0"}, StreakDay: 1}
}

func (f *fakeLedger) GetMonthSummary(context.Context, *connect.Request[moneyrecapv1.GetMonthSummaryRequest]) (*connect.Response[moneyrecapv1.GetMonthSummaryResponse], error) {
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&moneyrecapv1.GetMonthSummaryResponse{Summary: zeroMonth()}), nil
}

func (f *fakeLedger) GetSettings(context.Context, *connect.Request[moneyrecapv1.GetSettingsRequest]) (*connect.Response[moneyrecapv1.GetSettingsResponse], error) {
	f.settings++
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&moneyrecapv1.GetSettingsResponse{Settings: &moneyrecapv1.LedgerSettings{Salary: f.salary, SavingTarget: f.target}}), nil
}

func (f *fakeLedger) SaveTarget(_ context.Context, req *connect.Request[moneyrecapv1.SaveTargetRequest]) (*connect.Response[moneyrecapv1.SaveTargetResponse], error) {
	f.gotTarget, f.gotToken = req.Msg, req.Header().Get("Authorization")
	if f.err != nil {
		return nil, f.err
	}
	f.target = &moneyrecapv1.SavingTargetSetting{Percent: req.Msg.Percent, EffectiveFrom: "2026-09"}
	return connect.NewResponse(&moneyrecapv1.SaveTargetResponse{Settings: &moneyrecapv1.LedgerSettings{Salary: f.salary, SavingTarget: f.target}, Month: zeroMonth()}), nil
}

func (f *fakeLedger) SaveSalary(_ context.Context, req *connect.Request[moneyrecapv1.SaveSalaryRequest]) (*connect.Response[moneyrecapv1.SaveSalaryResponse], error) {
	f.gotSalary, f.gotToken = req.Msg, req.Header().Get("Authorization")
	if f.err != nil {
		return nil, f.err
	}
	f.salary = &moneyrecapv1.SalarySetting{Amount: req.Msg.Amount, Payday: req.Msg.Payday, EffectiveFrom: "2026-09", OriginalAmount: req.Msg.Amount}
	return connect.NewResponse(&moneyrecapv1.SaveSalaryResponse{Settings: &moneyrecapv1.LedgerSettings{Salary: f.salary}, Month: zeroMonth()}), nil
}

func (f *fakeLedger) RelabelCurrency(_ context.Context, req *connect.Request[moneyrecapv1.RelabelCurrencyRequest]) (*connect.Response[moneyrecapv1.RelabelCurrencyResponse], error) {
	f.gotRelabel = req.Msg
	if f.err != nil {
		return nil, f.err
	}
	if f.salary != nil {
		f.salary.Amount = &moneyrecapv1.Money{Amount: f.salary.Amount.Amount, CurrencyCode: req.Msg.CurrencyCode}
	}
	return connect.NewResponse(&moneyrecapv1.RelabelCurrencyResponse{Settings: &moneyrecapv1.LedgerSettings{Salary: f.salary, CurrencyCode: req.Msg.CurrencyCode}}), nil
}

// ledgerClient serves f the way the BFF reaches the real ledger: over Connect, forwarding the caller's token.
func ledgerClient(t *testing.T, f *fakeLedger) moneyrecapv1connect.LedgerServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewLedgerServiceHandler(f))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return moneyrecapv1connect.NewLedgerServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(auth.Forward()))
}
