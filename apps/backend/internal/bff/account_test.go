package bff

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

func (f *fakeMembers) RequestDeletion(_ context.Context, req *connect.Request[moneyrecapv1.RequestDeletionRequest]) (*connect.Response[moneyrecapv1.RequestDeletionResponse], error) {
	f.gotDeletion = req.Msg
	if f.cancelErr != nil {
		return nil, f.cancelErr
	}
	return connect.NewResponse(&moneyrecapv1.RequestDeletionResponse{Profile: &moneyrecapv1.Profile{MemberId: req.Msg.MemberId, DeletionRequestedAt: timestamppb.New(time.Unix(1_790_000_000, 0))}}), nil
}

// RequestDeletion refuses a keyless call before members, then hands members the key and the Member's id and answers
// members' stamp; a failure keeps its code, so the phone stays signed in (D11).
func TestRequestDeletionNeedsTheKeyAndAnswersTheStamp(t *testing.T) {
	members := &fakeMembers{}
	s := newService(t, members, &fakeNotify{})
	req := connect.NewRequest(&moneyrecapv1.BffServiceRequestDeletionRequest{})
	if _, err := s.RequestDeletion(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument || members.gotDeletion != nil {
		t.Fatalf("no key: %v, members reached %v; want invalid_argument and no call", err, members.gotDeletion != nil)
	}
	req.Header().Set("Idempotency-Key", "key-1")
	res, err := s.RequestDeletion(ctx, req)
	if err != nil || res.Msg.DeletionRequestedAt.AsTime().Unix() != 1_790_000_000 || members.gotDeletion.IdempotencyKey != "key-1" || members.gotDeletion.MemberId != "m-1" {
		t.Fatalf("RequestDeletion: %v, %v; members got %v", res, err, members.gotDeletion)
	}
	members.cancelErr = connect.NewError(connect.CodeUnavailable, errors.New("down"))
	if _, err := s.RequestDeletion(ctx, req); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("members down: %v, want unavailable", err)
	}
}

// ForgetDevice refuses a keyless call before notify, then hands notify the key, the Member's id and the token.
func TestForgetDeviceNeedsTheKey(t *testing.T) {
	notify := &fakeNotify{}
	s := newService(t, &fakeMembers{}, notify)
	req := connect.NewRequest(&moneyrecapv1.BffServiceForgetDeviceRequest{ApnsToken: "abc1"})
	if _, err := s.ForgetDevice(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument || notify.forgot != nil {
		t.Fatalf("no key: %v, notify reached %v; want invalid_argument and no call", err, notify.forgot != nil)
	}
	req.Header().Set("Idempotency-Key", "key-1")
	if _, err := s.ForgetDevice(ctx, req); err != nil || notify.forgot.GetIdempotencyKey() != "key-1" || notify.forgot.GetMemberId() != "m-1" || notify.forgot.GetApnsToken() != "abc1" {
		t.Fatalf("ForgetDevice: %v; notify got %v", err, notify.forgot)
	}
}
