package bff

import (
	"context"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// idempotencyKey is the phone's Idempotency-Key header, which every Member write needs (A14): a key
// minted here would make each retry a new write. RequestMagicLink mints one because it has no Member.
func idempotencyKey(h http.Header) (string, error) {
	key := h.Get("Idempotency-Key")
	if key == "" {
		return "", connect.NewError(connect.CodeInvalidArgument, errors.New("the Idempotency-Key header is required"))
	}
	return key, nil
}

// SetTimezone is the launch call (A3): it creates the Member on first sight, as GetHome does, then stores
// the zone under the phone's key. Its Profile carries the deletion stamp the restore sheet reads (D60).
func (s *Service) SetTimezone(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceSetTimezoneRequest]) (*connect.Response[moneyrecapv1.BffServiceSetTimezoneResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	set, err := s.Members.SetTimezone(ctx, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: key, MemberId: profile.MemberId, Timezone: req.Msg.Timezone}))
	if err != nil {
		return nil, requiredCallFailed("members.SetTimezone", connect.CodeOf(err), err)
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceSetTimezoneResponse{Profile: set.Msg.Profile}), nil
}

// CancelDeletion is "Keep my account" on the restore sheet (PRD §6.18, D60): members clears the stamp under the
// phone's key. A stamp past the 24 hours stays failed_precondition.
func (s *Service) CancelDeletion(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceCancelDeletionRequest]) (*connect.Response[moneyrecapv1.BffServiceCancelDeletionResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	kept, err := s.Members.CancelDeletion(ctx, connect.NewRequest(&moneyrecapv1.CancelDeletionRequest{IdempotencyKey: key, MemberId: profile.MemberId}))
	if err != nil {
		return nil, requiredCallFailed("members.CancelDeletion", connect.CodeOf(err), err)
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceCancelDeletionResponse{Profile: kept.Msg.Profile}), nil
}

// RegisterDevice hands the phone's push token to notify under the phone's key (PRD §10, I8). The phone ignores a
// failure, so a code here only reaches its log.
func (s *Service) RegisterDevice(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceRegisterDeviceRequest]) (*connect.Response[moneyrecapv1.BffServiceRegisterDeviceResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	if _, err := s.Notify.RegisterDevice(ctx, connect.NewRequest(&moneyrecapv1.RegisterDeviceRequest{IdempotencyKey: key, MemberId: profile.MemberId, ApnsToken: req.Msg.ApnsToken})); err != nil {
		return nil, requiredCallFailed("notify.RegisterDevice", connect.CodeOf(err), err)
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceRegisterDeviceResponse{}), nil
}

// LogSpend is the Log a spend sheet's Save (PRD §6.9, ARCHITECTURE §6.2): ledger writes the Spend under the phone's
// key, notify hears a current-month Spend's crossed_line, then the answer carries the current month's Home.
func (s *Service) LogSpend(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceLogSpendRequest]) (*connect.Response[moneyrecapv1.BffServiceLogSpendResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ledgerCtx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	logged, err := s.Ledger.LogSpend(ledgerCtx, connect.NewRequest(&moneyrecapv1.LogSpendRequest{
		IdempotencyKey: key, MemberId: profile.MemberId, Amount: req.Msg.Amount, Category: req.Msg.Category, Date: req.Msg.Date, Note: req.Msg.Note,
	}))
	if err != nil {
		return nil, requiredCallFailed("ledger.LogSpend", connect.CodeOf(err), err)
	}
	home, err := s.home(ctx, profile, s.evaluateWarning(ctx, key, profile, logged.Msg.Month, logged.Msg.CrossedLine))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceLogSpendResponse{Home: home, Spend: logged.Msg.Spend}), nil
}

// EditSpend is the edit sheet's Save changes (PRD §6.10): ledger rewrites the Spend under the phone's key, notify hears
// crossed_line when the Spend now sits in the current month, then the answer carries the current month's Home (§6.2).
func (s *Service) EditSpend(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceEditSpendRequest]) (*connect.Response[moneyrecapv1.BffServiceEditSpendResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ledgerCtx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	edited, err := s.Ledger.EditSpend(ledgerCtx, connect.NewRequest(&moneyrecapv1.EditSpendRequest{
		IdempotencyKey: key, MemberId: profile.MemberId, SpendId: req.Msg.SpendId, Amount: req.Msg.Amount, Category: req.Msg.Category, Date: req.Msg.Date, Note: req.Msg.Note,
	}))
	if err != nil {
		return nil, requiredCallFailed("ledger.EditSpend", connect.CodeOf(err), err)
	}
	home, err := s.home(ctx, profile, s.evaluateWarning(ctx, key, profile, edited.Msg.Month, edited.Msg.CrossedLine))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceEditSpendResponse{Home: home, Spend: edited.Msg.Spend}), nil
}

// DeleteSpend is the confirmed Delete this spend (PRD §6.10, D21): ledger stamps the Spend deleted under the phone's
// key. A delete never crosses the line, so notify is not told and Home's banner is GetWarningState's (§6.2).
func (s *Service) DeleteSpend(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceDeleteSpendRequest]) (*connect.Response[moneyrecapv1.BffServiceDeleteSpendResponse], error) {
	key, err := idempotencyKey(req.Header())
	if err != nil {
		return nil, err
	}
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	ledgerCtx, cancel := context.WithTimeout(ctx, s.timeout())
	defer cancel()
	if _, err := s.Ledger.DeleteSpend(ledgerCtx, connect.NewRequest(&moneyrecapv1.DeleteSpendRequest{IdempotencyKey: key, MemberId: profile.MemberId, SpendId: req.Msg.SpendId})); err != nil {
		return nil, requiredCallFailed("ledger.DeleteSpend", connect.CodeOf(err), err)
	}
	home, err := s.home(ctx, profile, nil)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceDeleteSpendResponse{Home: home}), nil
}

// evaluateTimeout bounds §6.2 step 4: a slow notify must not hold up a Spend already saved.
const evaluateTimeout = 300 * time.Millisecond

// evaluateWarning is §6.2 step 4: for a write in the current month, notify stamps it when ledger says the write
// crossed the line. nil for a past month or a failed call, and home then reads GetWarningState instead.
func (s *Service) evaluateWarning(ctx context.Context, key string, profile *moneyrecapv1.Profile, month *moneyrecapv1.MonthSummary, crossed bool) *moneyrecapv1.EvaluateWarningResponse {
	current := s.now().In(zone(profile)).Format(monthLayout)
	if month.GetMonth() != current {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, evaluateTimeout)
	defer cancel()
	res, err := s.Notify.EvaluateWarning(ctx, connect.NewRequest(&moneyrecapv1.EvaluateWarningRequest{
		IdempotencyKey: key, MemberId: profile.MemberId, Month: current, CrossedLine: crossed, Warning: month.GetTargetWarning(),
	}))
	if err != nil {
		logOptionalCallDegraded("notify.EvaluateWarning", "the banner read from GetWarningState", err)
		return nil
	}
	return res.Msg
}
