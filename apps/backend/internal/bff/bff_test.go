package bff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// fakeMembers is a MembersService that answers one fixed Member; createDelay and profileDelay slow
// GetOrCreateBySubject and GetProfile.
type fakeMembers struct {
	moneyrecapv1connect.UnimplementedMembersServiceHandler
	createDelay, profileDelay time.Duration
	gotToken                  string
}

func (f *fakeMembers) GetOrCreateBySubject(ctx context.Context, req *connect.Request[moneyrecapv1.GetOrCreateBySubjectRequest]) (*connect.Response[moneyrecapv1.GetOrCreateBySubjectResponse], error) {
	f.gotToken = req.Header().Get("Authorization")
	select {
	case <-time.After(f.createDelay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return connect.NewResponse(&moneyrecapv1.GetOrCreateBySubjectResponse{Profile: &moneyrecapv1.Profile{MemberId: "m-1", Email: req.Msg.Email}}), nil
}

func (f *fakeMembers) GetProfile(ctx context.Context, _ *connect.Request[moneyrecapv1.GetProfileRequest]) (*connect.Response[moneyrecapv1.GetProfileResponse], error) {
	select {
	case <-time.After(f.profileDelay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return connect.NewResponse(&moneyrecapv1.GetProfileResponse{Profile: &moneyrecapv1.Profile{MemberId: "m-1", Email: "test@example.com"}}), nil
}

// hangingNotify never answers, like a notify container that is stopped.
type hangingNotify struct{}

func (hangingNotify) GetWarningState(ctx context.Context, _ *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// firedNotify says the Target Warning fired and carries the §8.8 copy it kept (PRD §6.8).
type firedNotify struct{}

func (firedNotify) GetWarningState(context.Context, *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error) {
	return connect.NewResponse(&moneyrecapv1.GetWarningStateResponse{Fired: true, Warning: &moneyrecapv1.TargetWarning{Message: "You've passed your saving line."}}), nil
}

func newService(t *testing.T, members *fakeMembers, notify Notify) *Service {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewMembersServiceHandler(members))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Service{
		Members: moneyrecapv1connect.NewMembersServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(auth.Forward())),
		Ledger:  StubLedger{},
		Notify:  notify,
		Timeout: 100 * time.Millisecond,
		Now:     func() time.Time { return time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC) },
	}
}

var ctx = auth.NewContext(context.Background(), auth.Claims{Subject: "sub-1", Email: "test@example.com", Raw: "tok"})

// assertZeroHome is the HomeView of zeros both the unit test and the end-to-end test expect (#10):
// nothing spent or in, no Saved, Salary, target, warning or recaps, no name, the fallback currency, streak day 1.
func assertZeroHome(t *testing.T, h *moneyrecapv1.HomeView) {
	t.Helper()
	if h.Spent.Amount != "0" || h.Spent.CurrencyCode != fallbackCurrency || h.MoneyIn.Amount != "0" || h.Saved != nil ||
		h.Salary != nil || h.SavingTarget != nil || h.TargetWarning != nil || len(h.RecapMonths) != 0 ||
		h.DisplayName != "" || h.CurrencyCode != fallbackCurrency || h.StreakDay != 1 || h.DaysRemaining < 1 {
		t.Fatalf("HomeView: %v", h)
	}
}

func TestGetHomeZeros(t *testing.T) {
	members := &fakeMembers{}
	res, err := newService(t, members, StubNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	h := res.Msg.Home
	assertZeroHome(t, h)
	if h.Month != "2026-09" || h.DaysRemaining != 15 || h.FirstRecapOn != "2026-10-01" {
		t.Fatalf("HomeView dates: %v", h)
	}
	if members.gotToken != "Bearer tok" {
		t.Fatalf("members received %q, want the caller's token forwarded", members.gotToken)
	}
}

func TestGetHomeNotifyTimeoutStillAnswers(t *testing.T) {
	res, err := newService(t, &fakeMembers{}, hangingNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Msg.Home.TargetWarning != nil || res.Msg.Home.Spent.Amount != "0" {
		t.Fatalf("HomeView: %v", res.Msg.Home)
	}
}

func TestGetHomeBannerOnlyWhenNotifySaysFired(t *testing.T) {
	s := newService(t, &fakeMembers{}, StubNotify{})
	res, err := s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Msg.Home.TargetWarning != nil {
		t.Fatalf("not fired: %v", res.Msg.Home.TargetWarning)
	}
	s.Notify = firedNotify{}
	if res, err = s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{})); err != nil {
		t.Fatal(err)
	}
	if res.Msg.Home.TargetWarning.GetMessage() != "You've passed your saving line." {
		t.Fatalf("fired: %v", res.Msg.Home.TargetWarning)
	}
}

func TestGetHomeMembersTimeoutIsAnError(t *testing.T) {
	_, err := newService(t, &fakeMembers{profileDelay: time.Second}, StubNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if connect.CodeOf(err) != connect.CodeDeadlineExceeded {
		t.Fatalf("members timeout: %v, want deadline_exceeded", err)
	}
	t.Logf("members timeout: %v", err)
}

func TestGetHomeGetOrCreateTimeoutIsAnError(t *testing.T) {
	start := time.Now()
	_, err := newService(t, &fakeMembers{createDelay: 2 * time.Second}, StubNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if elapsed := time.Since(start); connect.CodeOf(err) != connect.CodeDeadlineExceeded || elapsed > time.Second {
		t.Fatalf("GetOrCreateBySubject timeout: %v after %v, want deadline_exceeded within the 100ms bound", err, elapsed)
	}
}

func TestDaysRemainingOnTheLastDay(t *testing.T) {
	s := newService(t, &fakeMembers{}, StubNotify{})
	s.Now = func() time.Time { return time.Date(2026, 2, 28, 23, 59, 0, 0, time.UTC) }
	res, err := s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if h := res.Msg.Home; h.DaysRemaining != 1 || h.FirstRecapOn != "2026-03-01" {
		t.Fatalf("days_remaining %d, first_recap_on %s", h.DaysRemaining, h.FirstRecapOn)
	}
}
