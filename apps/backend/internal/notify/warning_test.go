package notify

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// fakeMembers answers GetProfile for the one Member each bearer token owns, as members does (§4.2 point 3);
// the tests' tokens are the subject itself, and a nil map is members down.
type fakeMembers struct {
	moneyrecapv1connect.UnimplementedMembersServiceHandler
	owns map[string]string
}

func (f fakeMembers) GetProfile(_ context.Context, req *connect.Request[moneyrecapv1.GetProfileRequest]) (*connect.Response[moneyrecapv1.GetProfileResponse], error) {
	if f.owns == nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("members is down"))
	}
	if f.owns[strings.TrimPrefix(req.Header().Get("Authorization"), "Bearer ")] != req.Msg.MemberId {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	return connect.NewResponse(&moneyrecapv1.GetProfileResponse{Profile: &moneyrecapv1.Profile{MemberId: req.Msg.MemberId}}), nil
}

// warningService is a Service over the notify database whose members answers owns, with id's rows removed at the end.
func warningService(t *testing.T, id string, owns map[string]string) *Service {
	t.Helper()
	db := testDB(t)
	t.Cleanup(func() {
		for _, table := range []string{"member_facts", "month_flag"} {
			if _, err := db.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = $1`, id); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
	return &Service{DB: db, Members: membersClient(t, owns)}
}

// membersClient reaches a fakeMembers answering owns, forwarding the caller's token as the service does.
func membersClient(t *testing.T, owns map[string]string) moneyrecapv1connect.MembersServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewMembersServiceHandler(fakeMembers{owns: owns}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return moneyrecapv1connect.NewMembersServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(auth.Forward()))
}

func as(subject string) context.Context {
	return auth.NewContext(context.Background(), auth.Claims{Subject: subject, Raw: subject})
}

// The month is stamped once, by the first crossed_line, with the copy it carried; later calls, a repeated key
// included, answer that copy and never fired_now, and another month stays its own (PRD §8.8, A14).
func TestTheMonthFlagStampsOnceAndKeepsItsFirstWords(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000b001"
	s := warningService(t, id, map[string]string{"owner": id})
	ctx := as("owner")
	evaluate := func(key, month string, crossed bool, message string) (*moneyrecapv1.EvaluateWarningResponse, error) {
		req := &moneyrecapv1.EvaluateWarningRequest{IdempotencyKey: key, MemberId: id, Month: month, CrossedLine: crossed}
		if message != "" {
			req.Warning = &moneyrecapv1.TargetWarning{Message: message}
		}
		res, err := s.EvaluateWarning(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	state := func(month string) *moneyrecapv1.GetWarningStateResponse {
		t.Helper()
		res, err := s.GetWarningState(ctx, connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: id, Month: month}))
		if err != nil {
			t.Fatalf("GetWarningState %s: %v", month, err)
		}
		return res.Msg
	}
	first, later := "You've spent IDR 2.500.000 of the IDR 2.400.000 you meant to spend this month.", "You've passed your saving line. IDR 300.000 over."

	if got := state("2026-09"); got.Fired || got.Warning != nil {
		t.Fatalf("an unstamped month: %v, want not fired", got)
	}
	if got, err := evaluate("k0", "2026-09", false, ""); err != nil || got.Fired || got.FiredNow {
		t.Fatalf("crossed_line false: %v, %v, want nothing stamped", got, err)
	}
	if got, err := evaluate("k1", "2026-09", true, first); err != nil || !got.Fired || !got.FiredNow || got.Warning.GetMessage() != first {
		t.Fatalf("the first crossing: %v, %v, want fired now with its copy", got, err)
	}
	for _, key := range []string{"k1", "k2"} {
		if got, err := evaluate(key, "2026-09", true, later); err != nil || !got.Fired || got.FiredNow || got.Warning.GetMessage() != first {
			t.Errorf("%s after the stamp: %v, %v, want fired, not now, with the first words", key, got, err)
		}
	}
	if got, err := evaluate("k3", "2026-09", false, ""); err != nil || !got.Fired || got.Warning.GetMessage() != first {
		t.Errorf("crossed_line false after the stamp: %v, %v, want the kept copy", got, err)
	}
	if got := state("2026-09"); !got.Fired || got.Warning.GetMessage() != first {
		t.Errorf("GetWarningState after the stamp: %v, want the first words", got)
	}
	if got := state("2026-10"); got.Fired {
		t.Errorf("the next month: %v, want its own, unstamped", got)
	}
	for name, req := range map[string]*moneyrecapv1.EvaluateWarningRequest{
		"no key":           {MemberId: id, Month: "2026-09"},
		"a bad month":      {IdempotencyKey: "b", MemberId: id, Month: "2026-9"},
		"crossed, no copy": {IdempotencyKey: "b", MemberId: id, Month: "2026-09", CrossedLine: true},
	} {
		if _, err := s.EvaluateWarning(ctx, connect.NewRequest(req)); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", name, err)
		}
	}
	if _, err := s.GetWarningState(ctx, connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: id, Month: "September"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("GetWarningState of a bad month: %v, want invalid_argument", err)
	}
}

// A token that does not own the Member is not_found on both reads and stamps nothing, even after the owner filled
// the replica; a malformed id is not_found too, and members down is unavailable, never a pass (§4.2 point 3).
func TestAnotherMembersWarningIsNotFound(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000b002"
	s := warningService(t, id, map[string]string{"owner": id, "stranger": "00000000-0000-4000-8000-00000000b003"})
	stamp := func(ctx context.Context, memberID string) error {
		_, err := s.EvaluateWarning(ctx, connect.NewRequest(&moneyrecapv1.EvaluateWarningRequest{
			IdempotencyKey: "k", MemberId: memberID, Month: "2026-09", CrossedLine: true, Warning: &moneyrecapv1.TargetWarning{Message: "over"},
		}))
		return err
	}
	read := func(ctx context.Context, memberID string) error {
		_, err := s.GetWarningState(ctx, connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: memberID, Month: "2026-09"}))
		return err
	}

	if err := stamp(as("owner"), id); err != nil {
		t.Fatalf("the owner: %v", err)
	}
	var subject string
	if err := s.DB.QueryRow(t.Context(), `SELECT keycloak_subject FROM member_facts WHERE member_id = $1`, id).Scan(&subject); err != nil || subject != "owner" {
		t.Fatalf("the replica holds %q, %v, want the owner's subject", subject, err)
	}
	for name, call := range map[string]func(context.Context, string) error{"EvaluateWarning": stamp, "GetWarningState": read} {
		if err := call(as("stranger"), id); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("%s under another Member's token: %v, want not_found", name, err)
		}
		if err := call(as("owner"), "not-a-uuid"); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("%s of a malformed id: %v, want not_found", name, err)
		}
	}
	if err := read(as("owner"), id); err != nil {
		t.Errorf("the owner after the stranger: %v", err)
	}
	s.Members = membersClient(t, nil)
	if err := read(as("stranger"), id); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("a miss with members down: %v, want unavailable", err)
	}
}
