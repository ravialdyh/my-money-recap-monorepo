package bff

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

const setupDone = 4

var errConvertAfterSetup = errors.New("after onboarding a Currency change converts, through ChangeCurrency")

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
	home, err := s.home(ctx, p, nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceSaveSalaryResponse{Home: home, Settings: settingsView(p, saved.Msg.Settings)}), nil
}

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
	home, err := s.home(ctx, p, nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceSaveTargetResponse{Home: home, Settings: settingsView(p, saved.Msg.Settings)}), nil
}

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

func (s *Service) relabel(ctx context.Context, key string, before *moneyrecapv1.Profile, code string) (*moneyrecapv1.LedgerSettings, error) {
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
	return moved.Msg.Settings, nil
}
