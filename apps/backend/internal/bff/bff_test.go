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
	"google.golang.org/protobuf/proto"

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
	gotProfile  *moneyrecapv1.SaveProfileRequest
	gotCancel   *moneyrecapv1.CancelDeletionRequest
	cancelErr   error
}

func (f *fakeMembers) CancelDeletion(_ context.Context, req *connect.Request[moneyrecapv1.CancelDeletionRequest]) (*connect.Response[moneyrecapv1.CancelDeletionResponse], error) {
	f.gotCancel = req.Msg
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	return connect.NewResponse(&moneyrecapv1.CancelDeletionResponse{Profile: &moneyrecapv1.Profile{MemberId: req.Msg.MemberId}}), nil
}

// SaveProfile writes the answers sent over the stored profile and keeps the highest step, as members does.
func (f *fakeMembers) SaveProfile(_ context.Context, req *connect.Request[moneyrecapv1.SaveProfileRequest]) (*connect.Response[moneyrecapv1.SaveProfileResponse], error) {
	f.gotProfile = req.Msg
	p := &moneyrecapv1.Profile{MemberId: req.Msg.MemberId}
	if f.profile != nil {
		p = proto.CloneOf(f.profile)
	}
	if req.Msg.DisplayName != nil {
		p.DisplayName = *req.Msg.DisplayName
	}
	if req.Msg.AvatarKey != nil {
		p.AvatarKey = *req.Msg.AvatarKey
	}
	if req.Msg.CurrencyCode != nil {
		p.CurrencyCode = *req.Msg.CurrencyCode
	}
	p.OnboardingStep = max(p.OnboardingStep, req.Msg.GetOnboardingStep())
	return connect.NewResponse(&moneyrecapv1.SaveProfileResponse{Profile: p}), nil
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

func newService(t *testing.T, members *fakeMembers, notify Notify) *Service {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewMembersServiceHandler(members))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Service{
		Members: moneyrecapv1connect.NewMembersServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(auth.Forward())),
		Ledger:  ledgerClient(t, &fakeLedger{}),
		Notify:  notify,
		Timeout: 100 * time.Millisecond,
		Now:     func() time.Time { return time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC) },
	}
}

var ctx = auth.NewContext(context.Background(), auth.Claims{Subject: "sub-1", Email: "test@example.com", Raw: "tok"})

// onboarded is a Member who has answered steps 1 and 2, so GetHome answers them (spec #92).
func onboarded() *moneyrecapv1.Profile {
	return &moneyrecapv1.Profile{MemberId: "m-1", DisplayName: "Maya", AvatarKey: "fox", CurrencyCode: "IDR"}
}

// assertZeroHome is the HomeView of zeros the unit and end-to-end tests expect once steps 1 and 2 are answered:
// nothing spent or in, no Saved, Salary, target, warning or recaps, streak day 1, the name, every amount in Currency.
func assertZeroHome(t *testing.T, h *moneyrecapv1.HomeView, name, currency string) {
	t.Helper()
	if h.Spent.Amount != "0" || h.Spent.CurrencyCode != currency || h.MoneyIn.Amount != "0" || h.MoneyIn.CurrencyCode != currency ||
		h.Saved != nil || h.Salary != nil || h.SavingTarget != nil || h.TargetWarning != nil || len(h.RecapMonths) != 0 ||
		h.DisplayName != name || h.CurrencyCode != currency || h.StreakDay != 1 || h.DaysRemaining < 1 {
		t.Fatalf("HomeView: %v, want zeros for %s in %s", h, name, currency)
	}
}

func TestGetHomeZeros(t *testing.T) {
	members := &fakeMembers{profile: onboarded()}
	res, err := newService(t, members, &fakeNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	h := res.Msg.Home
	assertZeroHome(t, h, "Maya", "IDR")
	if h.Month != "2026-09" || h.DaysRemaining != 15 || h.FirstRecapOn != "2026-10-01" {
		t.Fatalf("HomeView dates: %v", h)
	}
	if members.gotToken != "Bearer tok" {
		t.Fatalf("members received %q, want the caller's token forwarded", members.gotToken)
	}
}

// GetHome refuses failed_precondition until steps 1 and 2 are answered, before it calls ledger or notify: Home
// greets by name and prints the Member's Currency, so neither has a fallback (spec #92).
func TestGetHomeRefusesBeforeStepsOneAndTwo(t *testing.T) {
	for name, p := range map[string]*moneyrecapv1.Profile{
		"a new Member": {MemberId: "m-1"},
		"no Avatar":    {MemberId: "m-1", DisplayName: "Maya", CurrencyCode: "IDR"},
		"no Currency":  {MemberId: "m-1", DisplayName: "Maya", AvatarKey: "fox"},
	} {
		s := newService(t, &fakeMembers{profile: p}, &fakeNotify{hang: true})
		s.Ledger = ledgerClient(t, &fakeLedger{err: errLedgerDown})
		if _, err := s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Errorf("%s: %v, want failed_precondition before the fan-out", name, err)
		}
	}
}

// GetHome builds the HomeView from the Profile GetOrCreateBySubject answers, with no second members read.
func TestGetHomeReadsTheProfileGetOrCreateReturned(t *testing.T) {
	members := &fakeMembers{profile: &moneyrecapv1.Profile{MemberId: "m-1", DisplayName: "Maya", AvatarKey: "fox", CurrencyCode: "IDR"}}
	res, err := newService(t, members, &fakeNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if h := res.Msg.Home; h.DisplayName != "Maya" || h.AvatarKey != "fox" || h.CurrencyCode != "IDR" || h.Spent.CurrencyCode != "IDR" {
		t.Fatalf("HomeView: %v", h)
	}
}

func TestGetHomeNotifyTimeoutStillAnswers(t *testing.T) {
	res, err := newService(t, &fakeMembers{profile: onboarded()}, &fakeNotify{hang: true}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Msg.Home.TargetWarning != nil || res.Msg.Home.Spent.Amount != "0" {
		t.Fatalf("HomeView: %v", res.Msg.Home)
	}
}

func TestGetHomeBannerOnlyWhenNotifySaysFired(t *testing.T) {
	s := newService(t, &fakeMembers{profile: onboarded()}, &fakeNotify{})
	res, err := s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Msg.Home.TargetWarning != nil {
		t.Fatalf("not fired: %v", res.Msg.Home.TargetWarning)
	}
	s.Notify = &fakeNotify{kept: &moneyrecapv1.TargetWarning{Message: "You've passed your saving line."}}
	if res, err = s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{})); err != nil {
		t.Fatal(err)
	}
	if res.Msg.Home.TargetWarning.GetMessage() != "You've passed your saving line." {
		t.Fatalf("fired: %v", res.Msg.Home.TargetWarning)
	}
}

func TestGetHomeGetOrCreateTimeoutIsAnError(t *testing.T) {
	start := time.Now()
	_, err := newService(t, &fakeMembers{createDelay: 2 * time.Second}, &fakeNotify{}).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if elapsed := time.Since(start); connect.CodeOf(err) != connect.CodeDeadlineExceeded || elapsed > 500*time.Millisecond {
		t.Fatalf("GetOrCreateBySubject timeout: %v after %v, want deadline_exceeded near the 100ms bound", err, elapsed)
	}
}

func TestGetHomeKeepsTheCodeAndDropsTheUpstreamText(t *testing.T) {
	const host = "172.18.0."
	cases := map[string]struct {
		members *fakeMembers
		ledger  *fakeLedger
		want    connect.Code
	}{
		"GetOrCreateBySubject": {&fakeMembers{createErr: connect.NewError(connect.CodeUnauthenticated, errors.New("fetching keys: dial tcp 172.18.0.2:8080: connection refused"))}, &fakeLedger{}, connect.CodeUnauthenticated},
		"GetMonthSummary":      {&fakeMembers{profile: onboarded()}, &fakeLedger{err: errLedgerDown}, connect.CodeUnavailable},
	}
	for name, c := range cases {
		s := newService(t, c.members, &fakeNotify{})
		s.Ledger = ledgerClient(t, c.ledger)
		_, err := s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
		if connect.CodeOf(err) != c.want || strings.Contains(err.Error(), host) {
			t.Errorf("%s: %v, want %v without the upstream text", name, err, c.want)
		}
	}
}

// Home's month is the Member's (A3): just past midnight on 1 October in Jakarta it is still 30 September in UTC,
// and ledger is asked for October, whose days left are ledger's figure, never the BFF's own clock.
func TestGetHomeReadsTheMonthInTheMembersZone(t *testing.T) {
	for zone, want := range map[string][2]string{"Asia/Jakarta": {"2026-10", "2026-11-01"}, "": {"2026-09", "2026-10-01"}} {
		profile := onboarded()
		profile.Timezone = zone
		ledger := &fakeLedger{}
		s := newService(t, &fakeMembers{profile: profile}, &fakeNotify{})
		s.Ledger = ledgerClient(t, ledger)
		s.Now = func() time.Time { return time.Date(2026, 9, 30, 18, 30, 0, 0, time.UTC) }
		res, err := s.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
		if err != nil {
			t.Fatal(err)
		}
		h := res.Msg.Home
		if ledger.gotMonth != want[0] || h.Month != want[0] || h.FirstRecapOn != want[1] || h.DaysRemaining != 15 {
			t.Errorf("zone %q: asked ledger for %s, Home %s, first recap %s, %d days left; want %s, %s, ledger's 15",
				zone, ledger.gotMonth, h.Month, h.FirstRecapOn, h.DaysRemaining, want[0], want[1])
		}
	}
}

// SetTimezone refuses a call without an Idempotency-Key before reaching members, and otherwise creates the
// Member first and hands members the phone's key, that Member's id and the zone (A14, #96).
func TestSetTimezoneNeedsTheKeyAndCreatesTheMemberFirst(t *testing.T) {
	members := &fakeMembers{}
	s := newService(t, members, &fakeNotify{})
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

// CancelDeletion refuses a keyless call before members, hands members the key and the Member's id, and keeps
// members' failed_precondition for a stamp past the 24 hours (D60, #105).
func TestCancelDeletionNeedsTheKeyAndKeepsTheCode(t *testing.T) {
	members := &fakeMembers{}
	s := newService(t, members, &fakeNotify{})
	req := connect.NewRequest(&moneyrecapv1.BffServiceCancelDeletionRequest{})
	if _, err := s.CancelDeletion(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument || members.gotCancel != nil {
		t.Fatalf("no key: %v, members reached %v; want invalid_argument and no call", err, members.gotCancel != nil)
	}
	req.Header().Set("Idempotency-Key", "key-1")
	if res, err := s.CancelDeletion(ctx, req); err != nil || res.Msg.Profile.MemberId != "m-1" || members.gotCancel.IdempotencyKey != "key-1" || members.gotCancel.MemberId != "m-1" {
		t.Fatalf("CancelDeletion: %v, %v; members got %v", res, err, members.gotCancel)
	}
	members.cancelErr = connect.NewError(connect.CodeFailedPrecondition, errors.New("the deletion request is over 24 hours old"))
	if _, err := s.CancelDeletion(ctx, req); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("a stamp past the window: %v, want failed_precondition", err)
	}
}

// RegisterDevice refuses a keyless call before notify, then hands notify the key, the Member's id and the token.
func TestRegisterDeviceNeedsTheKey(t *testing.T) {
	notify := &fakeNotify{}
	s := newService(t, &fakeMembers{}, notify)
	req := connect.NewRequest(&moneyrecapv1.BffServiceRegisterDeviceRequest{ApnsToken: "abc1"})
	if _, err := s.RegisterDevice(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument || notify.device != nil {
		t.Fatalf("no key: %v, notify reached %v; want invalid_argument and no call", err, notify.device != nil)
	}
	req.Header().Set("Idempotency-Key", "key-1")
	if _, err := s.RegisterDevice(ctx, req); err != nil || notify.device.GetIdempotencyKey() != "key-1" || notify.device.GetMemberId() != "m-1" || notify.device.GetApnsToken() != "abc1" {
		t.Fatalf("RegisterDevice: %v; notify got %v", err, notify.device)
	}
}
