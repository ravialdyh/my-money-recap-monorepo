// Package bff is the API the phone talks to (ARCHITECTURE §6): GetHome joins members, ledger and notify, GetRecap
// members and ledger; RequestMagicLink (magiclink.go) calls the realm and notify.
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

// Notify is the part of moneyrecapv1connect.NotifyServiceClient the Member paths need (§7.3): the Target Warning,
// the phone's push token, the notification preferences and the export email.
type Notify interface {
	GetWarningState(ctx context.Context, req *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error)
	EvaluateWarning(ctx context.Context, req *connect.Request[moneyrecapv1.EvaluateWarningRequest]) (*connect.Response[moneyrecapv1.EvaluateWarningResponse], error)
	RegisterDevice(ctx context.Context, req *connect.Request[moneyrecapv1.RegisterDeviceRequest]) (*connect.Response[moneyrecapv1.RegisterDeviceResponse], error)
	ForgetDevice(ctx context.Context, req *connect.Request[moneyrecapv1.ForgetDeviceRequest]) (*connect.Response[moneyrecapv1.ForgetDeviceResponse], error)
	GetPrefs(ctx context.Context, req *connect.Request[moneyrecapv1.GetPrefsRequest]) (*connect.Response[moneyrecapv1.GetPrefsResponse], error)
	SavePrefs(ctx context.Context, req *connect.Request[moneyrecapv1.SavePrefsRequest]) (*connect.Response[moneyrecapv1.SavePrefsResponse], error)
	EmailExport(ctx context.Context, req *connect.Request[moneyrecapv1.EmailExportRequest]) (*connect.Response[moneyrecapv1.EmailExportResponse], error)
}

// Service implements BffService. Every RPC but RequestMagicLink runs behind auth.Interceptor;
// Members and Ledger must carry auth.Forward so each service verifies the same token again.
type Service struct {
	moneyrecapv1connect.UnimplementedBffServiceHandler
	Members moneyrecapv1connect.MembersServiceClient
	Ledger  moneyrecapv1connect.LedgerServiceClient
	Notify  Notify
	// Currency quotes the day's rate; optional, so its absence is only the failed fetch (spec #169).
	Currency moneyrecapv1connect.CurrencyServiceClient
	// MagicLink is the realm and the notify client RequestMagicLink needs (magiclink.go, A10).
	MagicLink MagicLink
	// Timeout bounds each call to a service; zero means the §6.1 one second.
	Timeout time.Duration
	// Now is the server clock; nil means time.Now. GetHome reads it in the Member's zone (A3).
	Now func() time.Time
}

// errNotOnboarded is GetHome's answer until steps 1 and 2 are answered: Home greets by name and prints the
// Member's Currency (spec #92), and the phone reads it as "open onboarding".
var errNotOnboarded = errors.New("onboarding steps 1 and 2 are unanswered")

// The server's calendar strings (PRD §6.8): a Recap Month and a date.
const monthLayout, dateLayout = "2006-01", "2006-01-02"

// timeout is Timeout with its default applied.
func (s *Service) timeout() time.Duration {
	if s.Timeout == 0 {
		return time.Second
	}
	return s.Timeout
}

// now is the server clock in UTC: Now, or time.Now when it is nil.
func (s *Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}

// zone is the Member's IANA zone (A3), the one ledger reads "today" in, from the zone data linked in as members
// links it; UTC before the launch call has stored one.
func zone(p *moneyrecapv1.Profile) *time.Location {
	loc, err := time.LoadLocation(p.GetTimezone())
	if err != nil {
		return time.UTC
	}
	return loc
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

// GetRecap is one month's Recap view (PRD §6.11 to §6.14, §6.17): members for the Member, then ledger for the month,
// each required, each code passed through; the current Display Name and ledger's Currency, never a stored copy (§7).
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

// home is the current month's HomeView in the Member's zone: ledger's month, required, joined with notify's banner,
// optional, both asked at once. A write's EvaluateWarning answer, when there is one, is the banner instead (§6.2).
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

// homeView joins the Profile and ledger's month into the HomeView in the Currency ledger answered Spent in, stamping
// money left without a code with it. The month and first_recap_on are local dates, days_remaining ledger's (A2, A3).
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

// requiredCallFailed logs why a required call failed and answers with code and a message naming only
// the call, so no upstream text (a host, a database) reaches the phone; the callee's typed details, such as
// notify's EmailBounced, pass through.
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

// logOptionalCallDegraded records that an optional call failed and the request went on without it:
// Home without notify, a request the per-Member limit could not count. A log line stands in for the
// §4.4 and §6.1 metrics until OpenTelemetry lands (A11).
func logOptionalCallDegraded(call, consequence string, err error) {
	log.Printf("%s: %v (%s)", call, err, consequence)
}
