package bff

import (
	"context"
	"errors"
	"net/http"

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
