package bff

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// Steps 1 and 2 resume on their missing answers, steps 3 and 4 on the highest step saved (spec #92).
func TestNextStepFollowsTheAnswers(t *testing.T) {
	for want, p := range map[int32]*moneyrecapv1.Profile{
		1: {AvatarKey: "fox", CurrencyCode: "IDR", OnboardingStep: 4},
		2: {DisplayName: "Maya", AvatarKey: "fox", OnboardingStep: 1},
		3: {DisplayName: "Maya", AvatarKey: "fox", CurrencyCode: "IDR", OnboardingStep: 2},
		4: {DisplayName: "Maya", AvatarKey: "fox", CurrencyCode: "IDR", OnboardingStep: 3},
		0: {DisplayName: "Maya", AvatarKey: "fox", CurrencyCode: "IDR", OnboardingStep: 4},
	} {
		if got := nextStep(p); got != want {
			t.Errorf("nextStep(%v) = %d, want %d", p, got, want)
		}
	}
}

// The state carries the given_name the caller's own token holds, which fills step 1 in (PRD §6.4).
func TestGetOnboardingStateCarriesTheGivenName(t *testing.T) {
	s := newService(t, &fakeMembers{}, StubNotify{})
	named := auth.NewContext(context.Background(), auth.Claims{Subject: "sub-1", Email: "test@example.com", GivenName: "Test", Raw: "tok"})
	res, err := s.GetOnboardingState(named, connect.NewRequest(&moneyrecapv1.BffServiceGetOnboardingStateRequest{}))
	if err != nil || res.Msg.State.NextStep != 1 || res.Msg.State.SuggestedName != "Test" || res.Msg.State.Profile.MemberId != "m-1" {
		t.Fatalf("GetOnboardingState: %v, %v; want step 1, Test and the Member", res, err)
	}
}

// SaveProfile refuses a missing key or a bad answer before any members call, and otherwise hands members the
// phone's key, the Member's id and the answers as sent (A14).
func TestSaveProfileChecksBeforeAnyCall(t *testing.T) {
	members := &fakeMembers{}
	s := newService(t, members, StubNotify{})
	send := func(key string, msg *moneyrecapv1.BffServiceSaveProfileRequest) error {
		req := connect.NewRequest(msg)
		if key != "" {
			req.Header().Set("Idempotency-Key", key)
		}
		_, err := s.SaveProfile(ctx, req)
		return err
	}
	for name, err := range map[string]error{
		"no key":        send("", &moneyrecapv1.BffServiceSaveProfileRequest{AvatarKey: proto.String("fox")}),
		"a seventh pet": send("k", &moneyrecapv1.BffServiceSaveProfileRequest{AvatarKey: proto.String("cat")}),
		"a digit":       send("k", &moneyrecapv1.BffServiceSaveProfileRequest{DisplayName: proto.String("Maya2")}),
	} {
		if connect.CodeOf(err) != connect.CodeInvalidArgument || members.gotToken != "" || members.gotProfile != nil {
			t.Errorf("%s: %v, members reached: %v; want invalid_argument and no call", name, err, members.gotProfile != nil)
		}
	}
	if err := send("k1", &moneyrecapv1.BffServiceSaveProfileRequest{DisplayName: proto.String("Maya"), AvatarKey: proto.String("fox"), OnboardingStep: proto.Int32(1)}); err != nil {
		t.Fatal(err)
	}
	if got := members.gotProfile; got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.GetDisplayName() != "Maya" || got.GetAvatarKey() != "fox" || got.GetOnboardingStep() != 1 || got.CurrencyCode != nil {
		t.Fatalf("members got %v", got)
	}
}
