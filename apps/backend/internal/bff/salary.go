package bff

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// setupDone is onboarding_step at 4: step 4 saved or skipped, after which a Salary save is an edit (D56) and a
// Currency change converts rather than relabels (spec #92).
const setupDone = 4

// errConvertAfterSetup refuses a Currency change sent as a profile answer once onboarding is over.
var errConvertAfterSetup = errors.New("after onboarding a Currency change converts, through ChangeCurrency")

// SaveSalary is onboarding step 3 and, later, Settings' Salary (PRD §6.6, §6.16): ledger.SaveSalary under the phone's
// key, then, during setup, step 3 recorded in members after ledger answers. One call and one key for the phone.
func (s *Service) SaveSalary(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceSaveSalaryRequest]) (*connect.Response[moneyrecapv1.BffServiceSaveSalaryResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ledgerCtx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	saved, err := s.Ledger.SaveSalary(ledgerCtx, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{
		IdempotencyKey: key, MemberId: p.MemberId, Amount: req.Msg.Amount, Payday: req.Msg.Payday,
		AppliesFrom: req.Msg.AppliesFrom, Setup: p.OnboardingStep < setupDone,
	}))
	if err != nil {
		return nil, requiredCallFailed("ledger.SaveSalary", connect.CodeOf(err), err)
	}
	if p, err = s.recordStep(ctx, key, p, 3); err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceSaveSalaryResponse{
		Home:     homeView(p, saved.Msg.Month, nil, nil, s.now()),
		Settings: &moneyrecapv1.SettingsView{Profile: p, LedgerSettings: saved.Msg.Settings},
	}), nil
}

// SaveTarget is onboarding step 4 and, later, Settings' Saving Target (PRD §6.7, §6.16), as SaveSalary: ledger first,
// a setup save while onboarding is unfinished, then step 4 recorded in members under the same key.
func (s *Service) SaveTarget(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceSaveTargetRequest]) (*connect.Response[moneyrecapv1.BffServiceSaveTargetResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	p, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ledgerCtx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	saved, err := s.Ledger.SaveTarget(ledgerCtx, connect.NewRequest(&moneyrecapv1.SaveTargetRequest{
		IdempotencyKey: key, MemberId: p.MemberId, Percent: req.Msg.Percent, Setup: p.OnboardingStep < setupDone,
	}))
	if err != nil {
		return nil, requiredCallFailed("ledger.SaveTarget", connect.CodeOf(err), err)
	}
	if p, err = s.recordStep(ctx, key, p, setupDone); err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceSaveTargetResponse{
		Home:     homeView(p, saved.Msg.Month, nil, nil, s.now()),
		Settings: &moneyrecapv1.SettingsView{Profile: p, LedgerSettings: saved.Msg.Settings},
	}), nil
}

// recordStep saves step in members under the write's key once ledger has answered, when progress is below it;
// members keeps the highest step, so a lower one is never sent (spec #92).
func (s *Service) recordStep(ctx context.Context, key string, p *moneyrecapv1.Profile, step int32) (*moneyrecapv1.Profile, error) {
	if p.OnboardingStep >= step {
		return p, nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	saved, err := s.Members.SaveProfile(ctx, connect.NewRequest(&moneyrecapv1.SaveProfileRequest{
		IdempotencyKey: key, MemberId: p.MemberId, OnboardingStep: proto.Int32(step),
	}))
	if err != nil {
		return nil, requiredCallFailed("members.SaveProfile", connect.CodeOf(err), err)
	}
	return saved.Msg.Profile, nil
}

// relabel is the second half of a setup Currency change once members holds the new code: ledger moves any saved
// Salary to it, never converting (spec #92). A first Currency, before step 2 was ever saved, has nothing to move.
func (s *Service) relabel(ctx context.Context, key string, before *moneyrecapv1.Profile, code string) (*moneyrecapv1.SalarySetting, error) {
	if before.OnboardingStep < 2 {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	moved, err := s.Ledger.RelabelCurrency(ctx, connect.NewRequest(&moneyrecapv1.RelabelCurrencyRequest{
		IdempotencyKey: key, MemberId: before.MemberId, CurrencyCode: code,
	}))
	if err != nil {
		return nil, requiredCallFailed("ledger.RelabelCurrency", connect.CodeOf(err), err)
	}
	return moved.Msg.Settings.GetSalary(), nil
}
