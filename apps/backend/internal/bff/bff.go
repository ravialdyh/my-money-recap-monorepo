// Package bff is the API the phone talks to (ARCHITECTURE §6): GetHome joins members with the
// in-process ledger and notify stubs; RequestMagicLink (magiclink.go) calls the realm and the real notify.
package bff

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// Ledger is the part of moneyrecapv1connect.LedgerServiceClient GetHome needs (§7.2), StubLedger until
// the ledger service is built. Money with an empty currency code is stamped by GetHome.
type Ledger interface {
	GetMonthSummary(ctx context.Context, req *connect.Request[moneyrecapv1.GetMonthSummaryRequest]) (*connect.Response[moneyrecapv1.GetMonthSummaryResponse], error)
}

// Notify is the part of moneyrecapv1connect.NotifyServiceClient GetHome needs (§7.3), StubNotify until
// notify serves GetWarningState.
type Notify interface {
	GetWarningState(ctx context.Context, req *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error)
}

// StubLedger answers zeros: nothing spent, nothing in, no Saved, no Salary, no target, streak day 1.
type StubLedger struct{}

func (StubLedger) GetMonthSummary(context.Context, *connect.Request[moneyrecapv1.GetMonthSummaryRequest]) (*connect.Response[moneyrecapv1.GetMonthSummaryResponse], error) {
	return connect.NewResponse(&moneyrecapv1.GetMonthSummaryResponse{Summary: &moneyrecapv1.MonthSummary{
		Spent: &moneyrecapv1.Money{Amount: "0"}, MoneyIn: &moneyrecapv1.Money{Amount: "0"}, StreakDay: 1,
	}}), nil
}

// StubNotify answers "not fired".
type StubNotify struct{}

func (StubNotify) GetWarningState(context.Context, *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error) {
	return connect.NewResponse(&moneyrecapv1.GetWarningStateResponse{}), nil
}

// Service implements BffService. Every RPC but RequestMagicLink runs behind auth.Interceptor;
// Members must carry auth.Forward so the service verifies the same token again.
type Service struct {
	moneyrecapv1connect.UnimplementedBffServiceHandler
	Members moneyrecapv1connect.MembersServiceClient
	Ledger  Ledger
	Notify  Notify
	// MagicLink is the realm and the notify client RequestMagicLink needs (magiclink.go, A10).
	MagicLink MagicLink
	// Timeout bounds each call to a service; zero means the §6.1 one second.
	Timeout time.Duration
	// Now is the server clock; nil means time.Now. GetHome's dates are UTC until ledger computes them in
	// the Member's zone (A3).
	Now func() time.Time
}

// The fallback currency while a Member has none: the currency every artboard is drawn in.
// A skeleton-only state (#10): once onboarding saves a Currency, every Member has one.
const fallbackCurrency = "GBP"

// The server's calendar strings (PRD §6.8): a Recap Month and a date.
const monthLayout, dateLayout = "2006-01", "2006-01-02"

// timeout is Timeout with its default applied.
func (s *Service) timeout() time.Duration {
	if s.Timeout == 0 {
		return time.Second
	}
	return s.Timeout
}

// member creates the Member on first sight and answers their Profile: members.GetOrCreateBySubject with
// the token's subject and email claim (PRD §5 step 5), which forwards the caller's token unchanged.
func (s *Service) member(ctx context.Context) (*moneyrecapv1.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	claims := auth.From(ctx)
	created, err := s.Members.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: claims.Subject, Email: claims.Email}))
	if err != nil {
		return nil, requiredCallFailed("members.GetOrCreateBySubject", connect.CodeOf(err), err)
	}
	return created.Msg.Profile, nil
}

func (s *Service) GetHome(ctx context.Context, _ *connect.Request[moneyrecapv1.GetHomeRequest]) (*connect.Response[moneyrecapv1.GetHomeResponse], error) {
	nowFn, timeout := s.Now, s.timeout()
	if nowFn == nil {
		nowFn = time.Now
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	memberID := profile.MemberId
	now := nowFn().UTC()
	month := now.Format(monthLayout)

	var (
		wg                   sync.WaitGroup
		ledgerRes            *connect.Response[moneyrecapv1.GetMonthSummaryResponse]
		notifyRes            *connect.Response[moneyrecapv1.GetWarningStateResponse]
		ledgerErr, notifyErr error
	)
	call := func(f func(ctx context.Context)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			f(ctx)
		}()
	}
	call(func(ctx context.Context) {
		ledgerRes, ledgerErr = s.Ledger.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: memberID, Month: month}))
	})
	call(func(ctx context.Context) {
		notifyRes, notifyErr = s.Notify.GetWarningState(ctx, connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: memberID, Month: month}))
	})
	wg.Wait()
	if ledgerErr != nil {
		return nil, requiredCallFailed("ledger.GetMonthSummary", connect.CodeUnavailable, ledgerErr)
	}
	summary := ledgerRes.Msg.GetSummary()
	var warning *moneyrecapv1.TargetWarning
	if notifyErr != nil {
		logOptionalCallDegraded("notify.GetWarningState", "HomeView returned without a banner", notifyErr)
	} else if notifyRes.Msg.GetFired() {
		warning = notifyRes.Msg.GetWarning()
	}

	currency := profile.CurrencyCode
	if currency == "" {
		currency = fallbackCurrency
	}
	stamp := func(m *moneyrecapv1.Money) *moneyrecapv1.Money {
		if m != nil && m.CurrencyCode == "" {
			m.CurrencyCode = currency
		}
		return m
	}
	firstOfNext := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return connect.NewResponse(&moneyrecapv1.GetHomeResponse{Home: &moneyrecapv1.HomeView{
		DisplayName:   profile.DisplayName,
		AvatarKey:     profile.AvatarKey,
		CurrencyCode:  currency,
		Month:         month,
		Spent:         stamp(summary.GetSpent()),
		MoneyIn:       stamp(summary.GetMoneyIn()),
		Saved:         stamp(summary.GetSaved()),
		Salary:        summary.GetSalary(),
		SavingTarget:  summary.GetSavingTarget(),
		StreakDay:     summary.GetStreakDay(),
		DaysRemaining: daysRemainingInclusive(now, firstOfNext),
		TargetWarning: warning,
		RecapMonths:   ledgerRes.Msg.GetRecapMonths(),
		FirstRecapOn:  firstOfNext.Format(dateLayout),
	}}), nil
}

// requiredCallFailed logs why a required call failed and answers with code and a message naming only
// the call, so no upstream text (a host, a database) reaches the phone.
func requiredCallFailed(call string, code connect.Code, err error) error {
	log.Printf("%s: %v", call, err)
	return connect.NewError(code, fmt.Errorf("%s failed", call))
}

// logOptionalCallDegraded records that an optional fan-out call failed and the HomeView went out
// without it. A log line stands in for the §6.1 metric until OpenTelemetry lands (A11).
func logOptionalCallDegraded(call, consequence string, err error) {
	log.Printf("%s: %v (%s)", call, err, consequence)
}

// daysRemainingInclusive counts today as one, so the last day of the month answers 1 (PRD §8.7).
// Once ledger is built, GetHome reads MonthSummary.days_remaining, computed in the Member's zone (A2, A3).
func daysRemainingInclusive(now, firstOfNext time.Time) int32 {
	return int32(firstOfNext.AddDate(0, 0, -1).Day() - now.Day() + 1)
}
