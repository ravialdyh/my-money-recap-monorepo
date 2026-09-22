package bff

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// fakeMembers is a MembersService serving GetOrCreateBySubject alone, so a second members call fails
// GetHome. It answers profile, or a new Member with no answers; createDelay slows it, createErr fails it.
type fakeMembers struct {
	moneyrecapv1connect.UnimplementedMembersServiceHandler
	createDelay time.Duration
	createErr   error
	profile     *moneyrecapv1.Profile
	gotToken    string
	gotZone     *moneyrecapv1.SetTimezoneRequest
}

func (f *fakeMembers) SetTimezone(_ context.Context, req *connect.Request[moneyrecapv1.SetTimezoneRequest]) (*connect.Response[moneyrecapv1.SetTimezoneResponse], error) {
	f.gotZone = req.Msg
	return connect.NewResponse(&moneyrecapv1.SetTimezoneResponse{Profile: &moneyrecapv1.Profile{MemberId: req.Msg.MemberId, Timezone: req.Msg.Timezone}}), nil
}

func (f *fakeMembers) GetOrCreateBySubject(ctx context.Context, req *connect.Request[moneyrecapv1.GetOrCreateBySubjectRequest]) (*connect.Response[moneyrecapv1.GetOrCreateBySubjectResponse], error) {
	f.gotToken = req.Header().Get("Authorization")
	select {
	case <-time.After(f.createDelay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if f.createErr != nil {
		return nil, f.createErr
	}
	profile := f.profile
	if profile == nil {
		profile = &moneyrecapv1.Profile{MemberId: "m-1", Email: req.Msg.Email}
	}
	return connect.NewResponse(&moneyrecapv1.GetOrCreateBySubjectResponse{Profile: profile}), nil
}

// hangingNotify never answers, like a notify container that is stopped.
type hangingNotify struct{}

func (hangingNotify) GetWarningState(ctx context.Context, _ *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

// failingLedger fails the way a ledger whose database is gone would, naming the address.
type failingLedger struct{}

func (failingLedger) GetMonthSummary(context.Context, *connect.Request[moneyrecapv1.GetMonthSummaryRequest]) (*connect.Response[moneyrecapv1.GetMonthSummaryResponse], error) {
	return nil, errors.New("dial tcp 172.18.0.5:5432: connect: connection refused")
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

// GetHome builds the HomeView from the Profile GetOrCreateBySubject answers, with no second members read.
func TestGetHomeReadsTheProfileGetOrCreateReturned(t *testing.T) {
	members := &fakeMembers{profile: &moneyrecapv1.Profile{MemberId: "m-1", DisplayName: "Maya", AvatarKey: "fox", CurrencyCode: "IDR"}}
	res, err := newService(t, members, StubNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if h := res.Msg.Home; h.DisplayName != "Maya" || h.AvatarKey != "fox" || h.CurrencyCode != "IDR" || h.Spent.CurrencyCode != "IDR" {
		t.Fatalf("HomeView: %v", h)
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

func TestGetHomeGetOrCreateTimeoutIsAnError(t *testing.T) {
	start := time.Now()
	_, err := newService(t, &fakeMembers{createDelay: 2 * time.Second}, StubNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if elapsed := time.Since(start); connect.CodeOf(err) != connect.CodeDeadlineExceeded || elapsed > 500*time.Millisecond {
		t.Fatalf("GetOrCreateBySubject timeout: %v after %v, want deadline_exceeded near the 100ms bound", err, elapsed)
	}
}

func TestGetHomeKeepsTheCodeAndDropsTheUpstreamText(t *testing.T) {
	const host = "172.18.0."
	cases := map[string]struct {
		members *fakeMembers
		ledger  Ledger
		want    connect.Code
	}{
		"GetOrCreateBySubject": {&fakeMembers{createErr: connect.NewError(connect.CodeUnauthenticated, errors.New("fetching keys: dial tcp 172.18.0.2:8080: connection refused"))}, StubLedger{}, connect.CodeUnauthenticated},
		"GetMonthSummary":      {&fakeMembers{}, failingLedger{}, connect.CodeUnavailable},
	}
	for name, c := range cases {
		s := newService(t, c.members, StubNotify{})
		s.Ledger = c.ledger
		_, err := s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
		if connect.CodeOf(err) != c.want || strings.Contains(err.Error(), host) {
			t.Errorf("%s: %v, want %v without the upstream text", name, err, c.want)
		}
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

// SetTimezone refuses a call without an Idempotency-Key before reaching members, and otherwise creates the
// Member first and hands members the phone's key, that Member's id and the zone (A14, #96).
func TestSetTimezoneNeedsTheKeyAndCreatesTheMemberFirst(t *testing.T) {
	members := &fakeMembers{}
	s := newService(t, members, StubNotify{})
	req := connect.NewRequest(&moneyrecapv1.BffServiceSetTimezoneRequest{Timezone: "Asia/Jakarta"})
	if _, err := s.SetTimezone(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument || members.gotToken != "" || members.gotZone != nil {
		t.Fatalf("no key: %v, members reached: %v; want invalid_argument and no members call", err, members.gotToken != "" || members.gotZone != nil)
	}
	req.Header().Set("Idempotency-Key", "key-1")
	res, err := s.SetTimezone(ctx, req)
	if err != nil || res.Msg.Profile.Timezone != "Asia/Jakarta" {
		t.Fatalf("SetTimezone: %v, %v", res, err)
	}
	if got := members.gotZone; members.gotToken != "Bearer tok" || got.IdempotencyKey != "key-1" || got.MemberId != "m-1" || got.Timezone != "Asia/Jakarta" {
		t.Fatalf("members got token %q and %v, want the caller's token, key-1, m-1 and Asia/Jakarta", members.gotToken, got)
	}
}
