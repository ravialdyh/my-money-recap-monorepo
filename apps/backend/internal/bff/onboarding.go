package bff

import (
	"context"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/answers"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// GetOnboardingState is where onboarding resumes (PRD §6.4): the first unanswered step, the answers saved so far,
// step 3's from ledger once steps 3 or 4 are next, and the given_name the Member's own token carries for step 1.
func (s *Service) GetOnboardingState(ctx context.Context, _ *connect.Request[moneyrecapv1.BffServiceGetOnboardingStateRequest]) (*connect.Response[moneyrecapv1.BffServiceGetOnboardingStateResponse], error) {
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	state := &moneyrecapv1.OnboardingState{NextStep: nextStep(p), Profile: p, SuggestedName: auth.From(ctx).GivenName}
	if state.NextStep == 3 || state.NextStep == 4 {
		ctx, cancel := context.WithTimeout(ctx, s.timeout())
		defer cancel()
		settings, err := s.Ledger.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: p.MemberId}))
		if err != nil {
			return nil, requiredCallFailed("ledger.GetSettings", connect.CodeUnavailable, err)
		}
		state.Salary = settings.Msg.Settings.GetSalary()
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceGetOnboardingStateResponse{State: state}), nil
}

// nextStep is the first unanswered step, 0 when none is: steps 1 and 2 by their answers, steps 3 and 4, which
// can be skipped, by the highest step saved, never by the progress number alone (spec #92).
func nextStep(p *moneyrecapv1.Profile) int32 {
	switch {
	case p.DisplayName == "" || p.AvatarKey == "":
		return 1
	case p.CurrencyCode == "":
		return 2
	case p.OnboardingStep < 3:
		return 3
	case p.OnboardingStep < 4:
		return 4
	}
	return 0
}

// SaveProfile checks the answers (internal/answers), then saves them in members under the phone's key (A14). A
// Currency saved during setup relabels step 3's Salary in ledger; after setup a change is refused (spec #92).
func (s *Service) SaveProfile(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceSaveProfileRequest]) (*connect.Response[moneyrecapv1.BffServiceSaveProfileResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	msg := req.Msg
	if _, err := answers.Check(msg.DisplayName, msg.AvatarKey, msg.CurrencyCode, msg.OnboardingStep); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	if msg.CurrencyCode != nil && *msg.CurrencyCode != p.CurrencyCode && p.OnboardingStep >= setupDone {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errConvertAfterSetup)
	}
	membersCtx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	saved, err := s.Members.SaveProfile(membersCtx, connect.NewRequest(&moneyrecapv1.SaveProfileRequest{
		IdempotencyKey: key, MemberId: p.MemberId, DisplayName: msg.DisplayName, AvatarKey: msg.AvatarKey,
		CurrencyCode: msg.CurrencyCode, OnboardingStep: msg.OnboardingStep,
	}))
	if err != nil {
		return nil, requiredCallFailed("members.SaveProfile", connect.CodeOf(err), err)
	}
	answer := &moneyrecapv1.BffServiceSaveProfileResponse{Profile: saved.Msg.Profile}
	if msg.CurrencyCode != nil && p.OnboardingStep < setupDone {
		if answer.Salary, err = s.relabel(ctx, key, p, *msg.CurrencyCode); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(answer), nil
}
