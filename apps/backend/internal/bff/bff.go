// Package bff is the API the phone talks to (ARCHITECTURE §6). Slice 1 (#17) answers GetHome
// with a HomeView of zeros: members is real, ledger and notify are in-process stubs.
package bff

import (
	"context"
	"log"
	"sync"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// MonthSummary is what ledger.GetMonthSummary will answer (§7.2): every figure Home shows for
// one month. Money fields with an empty currency code are stamped with the Member's currency.
type MonthSummary struct {
	Spent, MoneyIn *moneyrecapv1.Money
	Saved          *moneyrecapv1.Money
	Salary         *moneyrecapv1.SalaryFacts
	SavingTarget   *moneyrecapv1.SavingTargetFacts
	StreakDay      int32
	RecapMonths    []*moneyrecapv1.RecapMonth
}

// Ledger is the slice of §7.2 that GetHome needs; a Connect client replaces the stub in slice 5.
type Ledger interface {
	GetMonthSummary(ctx context.Context, memberID, month string) (MonthSummary, error)
}

// Notify is the slice of §7.3 that GetHome needs; a Connect client replaces the stub in slice 7.
type Notify interface {
	GetWarningState(ctx context.Context, memberID, month string) (*moneyrecapv1.TargetWarning, error)
}

// StubLedger answers zeros: nothing spent, nothing in, no Saved, no Salary, no target, streak day 1.
type StubLedger struct{}

func (StubLedger) GetMonthSummary(context.Context, string, string) (MonthSummary, error) {
	return MonthSummary{Spent: &moneyrecapv1.Money{Amount: "0"}, MoneyIn: &moneyrecapv1.Money{Amount: "0"}, StreakDay: 1}, nil
}

// StubNotify answers "no warning".
type StubNotify struct{}

func (StubNotify) GetWarningState(context.Context, string, string) (*moneyrecapv1.TargetWarning, error) {
	return nil, nil
}

// Service implements BffService. Every RPC runs behind auth.Interceptor; Members must carry
// auth.Forward so the service verifies the same token again.
type Service struct {
	moneyrecapv1connect.UnimplementedBffServiceHandler
	Members moneyrecapv1connect.MembersServiceClient
	Ledger  Ledger
	Notify  Notify
	// Timeout bounds each fan-out call; zero means the §6.1 one second.
	Timeout time.Duration
	// Now is the server clock; nil means time.Now. UTC until SetTimezone lands (#10, A3).
	Now func() time.Time
}

// The fallback currency while a Member has none: the currency every artboard is drawn in.
// A skeleton-only state; slice 4 removes it (#10).
const fallbackCurrency = "GBP"

// The server's calendar strings (PRD §6.8): a Recap Month and a date.
const monthLayout, dateLayout = "2006-01", "2006-01-02"

func (s *Service) GetHome(ctx context.Context, _ *connect.Request[moneyrecapv1.GetHomeRequest]) (*connect.Response[moneyrecapv1.GetHomeResponse], error) {
	claims := auth.From(ctx)
	created, err := s.Members.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: claims.Subject, Email: claims.Email}))
	if err != nil {
		return nil, err
	}
	memberID := created.Msg.Profile.MemberId
	nowFn, timeout := s.Now, s.Timeout
	if nowFn == nil {
		nowFn = time.Now
	}
	if timeout == 0 {
		timeout = time.Second
	}
	now := nowFn().UTC()
	month := now.Format(monthLayout)

	var (
		wg                       sync.WaitGroup
		profile                  *moneyrecapv1.Profile
		summary                  MonthSummary
		warning                  *moneyrecapv1.TargetWarning
		profErr, sumErr, warnErr error
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
		var res *connect.Response[moneyrecapv1.GetProfileResponse]
		if res, profErr = s.Members.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: memberID})); profErr == nil {
			profile = res.Msg.Profile
		}
	})
	call(func(ctx context.Context) { summary, sumErr = s.Ledger.GetMonthSummary(ctx, memberID, month) })
	call(func(ctx context.Context) { warning, warnErr = s.Notify.GetWarningState(ctx, memberID, month) })
	wg.Wait()
	if profErr != nil {
		return nil, profErr
	}
	if sumErr != nil {
		return nil, connect.NewError(connect.CodeUnavailable, sumErr)
	}
	if warnErr != nil {
		logOptionalCallDegraded("notify.GetWarningState", "HomeView returned without a banner", warnErr)
		warning = nil
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
		Spent:         stamp(summary.Spent),
		MoneyIn:       stamp(summary.MoneyIn),
		Saved:         stamp(summary.Saved),
		Salary:        summary.Salary,
		SavingTarget:  summary.SavingTarget,
		StreakDay:     summary.StreakDay,
		DaysRemaining: daysRemainingInclusive(now, firstOfNext),
		TargetWarning: warning,
		RecapMonths:   summary.RecapMonths,
		FirstRecapOn:  firstOfNext.Format(dateLayout),
	}}), nil
}

// logOptionalCallDegraded records that an optional fan-out call failed and the HomeView went out
// without it. A log line stands in for the §6.1 metric until OpenTelemetry lands (slice 8, A11).
func logOptionalCallDegraded(call, consequence string, err error) {
	log.Printf("%s: %v (%s)", call, err, consequence)
}

// daysRemainingInclusive counts today as one, so the last day of the month answers 1 (PRD §8.7).
func daysRemainingInclusive(now, firstOfNext time.Time) int32 {
	return int32(firstOfNext.AddDate(0, 0, -1).Day() - now.Day() + 1)
}
