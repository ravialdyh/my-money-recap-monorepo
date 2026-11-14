package bff

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	_ "time/tzdata"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

type Notify interface {
	GetWarningState(ctx context.Context, req *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error)
	EvaluateWarning(ctx context.Context, req *connect.Request[moneyrecapv1.EvaluateWarningRequest]) (*connect.Response[moneyrecapv1.EvaluateWarningResponse], error)
	RegisterDevice(ctx context.Context, req *connect.Request[moneyrecapv1.RegisterDeviceRequest]) (*connect.Response[moneyrecapv1.RegisterDeviceResponse], error)
	ForgetDevice(ctx context.Context, req *connect.Request[moneyrecapv1.ForgetDeviceRequest]) (*connect.Response[moneyrecapv1.ForgetDeviceResponse], error)
	GetPrefs(ctx context.Context, req *connect.Request[moneyrecapv1.GetPrefsRequest]) (*connect.Response[moneyrecapv1.GetPrefsResponse], error)
	SavePrefs(ctx context.Context, req *connect.Request[moneyrecapv1.SavePrefsRequest]) (*connect.Response[moneyrecapv1.SavePrefsResponse], error)
	EmailExport(ctx context.Context, req *connect.Request[moneyrecapv1.EmailExportRequest]) (*connect.Response[moneyrecapv1.EmailExportResponse], error)
}

type Service struct {
	moneyrecapv1connect.UnimplementedBffServiceHandler
	Members   moneyrecapv1connect.MembersServiceClient
	Ledger    moneyrecapv1connect.LedgerServiceClient
	Notify    Notify
	Currency  moneyrecapv1connect.CurrencyServiceClient
	MagicLink MagicLink
	Timeout   time.Duration
	Now       func() time.Time
}

var errNotOnboarded = errors.New("onboarding steps 1 and 2 are unanswered")

const monthLayout, dateLayout = "2006-01", "2006-01-02"

func (s *Service) timeout() time.Duration {
	if s.Timeout == 0 {
		return time.Second
	}
	return s.Timeout
}

func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}

func zone(p *moneyrecapv1.Profile) *time.Location {
	loc, err := time.LoadLocation(p.GetTimezone())
	if err != nil {
		return time.UTC
	}
	return loc
}

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
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	if step := nextStep(profile); step == 1 || step == 2 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errNotOnboarded)
	}
	home, err := s.home(ctx, profile, nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.GetHomeResponse{Home: home}), nil
}

func (s *Service) GetRecap(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceGetRecapRequest]) (*connect.Response[moneyrecapv1.BffServiceGetRecapResponse], error) {
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	res, err := s.Ledger.GetRecap(ctx, connect.NewRequest(&moneyrecapv1.GetRecapRequest{MemberId: profile.MemberId, Month: req.Msg.Month}))
	if err != nil {
		return nil, requiredCallFailed("ledger.GetRecap", connect.CodeOf(err), err)
	}
	code, _ := heldIn(res.Msg.Recap.GetSpent().GetCurrencyCode(), profile)
	return connect.NewResponse(&moneyrecapv1.BffServiceGetRecapResponse{Recap: &moneyrecapv1.RecapView{
		Recap: res.Msg.Recap, DisplayName: profile.DisplayName, CurrencyCode: code,
	}}), nil
}

func (s *Service) home(ctx context.Context, profile *moneyrecapv1.Profile, evaluated *moneyrecapv1.EvaluateWarningResponse) (*moneyrecapv1.HomeView, error) {
	timeout := s.timeout()
	memberID := profile.MemberId
	now := s.now().In(zone(profile))
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
	if evaluated == nil {
		call(func(ctx context.Context) {
			notifyRes, notifyErr = s.Notify.GetWarningState(ctx, connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: memberID, Month: month}))
		})
	}
	wg.Wait()
	if ledgerErr != nil {
		return nil, requiredCallFailed("ledger.GetMonthSummary", connect.CodeUnavailable, ledgerErr)
	}
	var warning *moneyrecapv1.TargetWarning
	switch {
	case evaluated != nil:
		if evaluated.GetFired() {
			warning = evaluated.GetWarning()
		}
	case notifyErr != nil:
		logOptionalCallDegraded("notify.GetWarningState", "HomeView returned without a banner", notifyErr)
	case notifyRes.Msg.GetFired():
		warning = notifyRes.Msg.GetWarning()
	}
	home := homeView(profile, ledgerRes.Msg.GetSummary(), ledgerRes.Msg.GetRecapMonths(), warning, now)
	home.EarliestSpendDate = ledgerRes.Msg.GetEarliestSpendDate()
	return home, nil
}

func homeView(p *moneyrecapv1.Profile, summary *moneyrecapv1.MonthSummary, recaps []*moneyrecapv1.RecapMonth, warning *moneyrecapv1.TargetWarning, now time.Time) *moneyrecapv1.HomeView {
	code, digits := heldIn(summary.GetSpent().GetCurrencyCode(), p)
	stamp := func(m *moneyrecapv1.Money) *moneyrecapv1.Money {
		if m != nil && m.CurrencyCode == "" {
			m.CurrencyCode = code
		}
		return m
	}
	firstOfNext := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	return &moneyrecapv1.HomeView{
		DisplayName:    p.DisplayName,
		AvatarKey:      p.AvatarKey,
		CurrencyCode:   code,
		CurrencyDigits: digits,
		Month:          now.Format(monthLayout),
		Spent:          stamp(summary.GetSpent()),
		MoneyIn:        stamp(summary.GetMoneyIn()),
		Saved:          stamp(summary.GetSaved()),
		Salary:         summary.GetSalary(),
		SavingTarget:   summary.GetSavingTarget(),
		StreakDay:      summary.GetStreakDay(),
		DaysRemaining:  summary.GetDaysRemaining(),
		TargetWarning:  warning,
		RecapMonths:    recaps,
		FirstRecapOn:   firstOfNext.Format(dateLayout),
	}
}

func requiredCallFailed(call string, code connect.Code, err error) error {
	log.Printf("%s: %v", call, err)
	failed := connect.NewError(code, fmt.Errorf("%s failed", call))
	var ce *connect.Error
	if errors.As(err, &ce) {
		for _, d := range ce.Details() {
			failed.AddDetail(d)
		}
	}
	return failed
}

func logOptionalCallDegraded(call, consequence string, err error) {
	log.Printf("%s: %v (%s)", call, err, consequence)
}
