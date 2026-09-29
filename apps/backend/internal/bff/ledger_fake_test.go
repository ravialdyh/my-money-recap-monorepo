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

// fakeLedger is ledger as the BFF sees it: a zero month with 15 days left and recaps, settings of salary, target and
// code, each read and write recorded with its token; err fails every call. LogSpend and EditSpend answer warning and crossed.
type fakeLedger struct {
	moneyrecapv1connect.UnimplementedLedgerServiceHandler
	err        error
	salary     *moneyrecapv1.SalarySetting
	target     *moneyrecapv1.SavingTargetSetting
	gotToken   string
	gotSalary  *moneyrecapv1.SaveSalaryRequest
	gotTarget  *moneyrecapv1.SaveTargetRequest
	gotRelabel *moneyrecapv1.RelabelCurrencyRequest
	gotSpend   *moneyrecapv1.LogSpendRequest
	gotEdit    *moneyrecapv1.EditSpendRequest
	gotDelete  *moneyrecapv1.DeleteSpendRequest
	crossed    bool
	warning    *moneyrecapv1.TargetWarning
	settings   int
	gotMonth   string
	gotList    string
	code       string
	recaps     []*moneyrecapv1.RecapMonth
}

func zeroMonth() *moneyrecapv1.MonthSummary {
	return &moneyrecapv1.MonthSummary{Spent: &moneyrecapv1.Money{Amount: "0"}, MoneyIn: &moneyrecapv1.Money{Amount: "0"}, StreakDay: 1}
}

func (f *fakeLedger) GetMonthSummary(_ context.Context, req *connect.Request[moneyrecapv1.GetMonthSummaryRequest]) (*connect.Response[moneyrecapv1.GetMonthSummaryResponse], error) {
	f.gotMonth = req.Msg.Month
	if f.err != nil {
		return nil, f.err
	}
	summary := zeroMonth()
	summary.DaysRemaining = 15
	return connect.NewResponse(&moneyrecapv1.GetMonthSummaryResponse{Summary: summary, RecapMonths: f.recaps, EarliestSpendDate: "2026-07-19"}), nil
}

func (f *fakeLedger) GetSettings(context.Context, *connect.Request[moneyrecapv1.GetSettingsRequest]) (*connect.Response[moneyrecapv1.GetSettingsResponse], error) {
	f.settings++
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&moneyrecapv1.GetSettingsResponse{Settings: &moneyrecapv1.LedgerSettings{Salary: f.salary, SavingTarget: f.target, CurrencyCode: f.code}}), nil
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

func (f *fakeLedger) LogSpend(_ context.Context, req *connect.Request[moneyrecapv1.LogSpendRequest]) (*connect.Response[moneyrecapv1.LogSpendResponse], error) {
	f.gotSpend, f.gotToken = req.Msg, req.Header().Get("Authorization")
	if f.err != nil {
		return nil, f.err
	}
	month := zeroMonth()
	month.Month, month.TargetWarning = req.Msg.Date[:7], f.warning
	return connect.NewResponse(&moneyrecapv1.LogSpendResponse{Spend: &moneyrecapv1.Spend{Id: "s-1", Amount: req.Msg.Amount, Date: req.Msg.Date}, Month: month, CrossedLine: f.crossed}), nil
}

func (f *fakeLedger) EditSpend(_ context.Context, req *connect.Request[moneyrecapv1.EditSpendRequest]) (*connect.Response[moneyrecapv1.EditSpendResponse], error) {
	f.gotEdit, f.gotToken = req.Msg, req.Header().Get("Authorization")
	if f.err != nil {
		return nil, f.err
	}
	month := zeroMonth()
	month.Month, month.TargetWarning = req.Msg.Date[:7], f.warning
	return connect.NewResponse(&moneyrecapv1.EditSpendResponse{Spend: &moneyrecapv1.Spend{Id: req.Msg.SpendId, Amount: req.Msg.Amount, Date: req.Msg.Date}, Month: month, CrossedLine: f.crossed}), nil
}

func (f *fakeLedger) DeleteSpend(_ context.Context, req *connect.Request[moneyrecapv1.DeleteSpendRequest]) (*connect.Response[moneyrecapv1.DeleteSpendResponse], error) {
	f.gotDelete, f.gotToken = req.Msg, req.Header().Get("Authorization")
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&moneyrecapv1.DeleteSpendResponse{Spend: &moneyrecapv1.Spend{Id: req.Msg.SpendId}, Month: zeroMonth()}), nil
}

func (f *fakeLedger) ListSpends(_ context.Context, req *connect.Request[moneyrecapv1.ListSpendsRequest]) (*connect.Response[moneyrecapv1.ListSpendsResponse], error) {
	f.gotList, f.gotToken = req.Msg.Month, req.Header().Get("Authorization")
	if f.err != nil {
		return nil, f.err
	}
	spend := &moneyrecapv1.Spend{Id: "s-1", Amount: &moneyrecapv1.Money{Amount: "25000", CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS}
	return connect.NewResponse(&moneyrecapv1.ListSpendsResponse{Spends: []*moneyrecapv1.Spend{spend}, Spent: spend.Amount, Categories: []*moneyrecapv1.CategoryShare{
		{Category: spend.Category, Amount: spend.Amount, Percent: 100, SpendCount: 1},
	}}), nil
}

func (f *fakeLedger) ListMonthsWithData(context.Context, *connect.Request[moneyrecapv1.ListMonthsWithDataRequest]) (*connect.Response[moneyrecapv1.ListMonthsWithDataResponse], error) {
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&moneyrecapv1.ListMonthsWithDataResponse{Months: []string{"2026-10", "2026-09"}}), nil
}

func (f *fakeLedger) GetRecap(_ context.Context, req *connect.Request[moneyrecapv1.GetRecapRequest]) (*connect.Response[moneyrecapv1.GetRecapResponse], error) {
	f.gotMonth, f.gotToken = req.Msg.Month, req.Header().Get("Authorization")
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(&moneyrecapv1.GetRecapResponse{Recap: &moneyrecapv1.Recap{Month: req.Msg.Month, ArchetypeTitle: "The Squirrel"}}), nil
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
