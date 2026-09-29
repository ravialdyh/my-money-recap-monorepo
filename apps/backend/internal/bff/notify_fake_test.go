package bff

import (
	"context"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// fakeNotify is notify's Target Warning as the BFF sees it: kept is the stamped copy, nil while unstamped, and
// EvaluateWarning stamps it on a crossed_line as notify does; hang never answers, like a stopped container.
type fakeNotify struct {
	hang      bool
	kept      *moneyrecapv1.TargetWarning
	evaluated *moneyrecapv1.EvaluateWarningRequest
	device    *moneyrecapv1.RegisterDeviceRequest
	reads     int
}

func (f *fakeNotify) GetWarningState(ctx context.Context, _ *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error) {
	f.reads++
	if f.hang {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return connect.NewResponse(&moneyrecapv1.GetWarningStateResponse{Fired: f.kept != nil, Warning: f.kept}), nil
}

func (f *fakeNotify) EvaluateWarning(ctx context.Context, req *connect.Request[moneyrecapv1.EvaluateWarningRequest]) (*connect.Response[moneyrecapv1.EvaluateWarningResponse], error) {
	f.evaluated = req.Msg
	if f.hang {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	now := req.Msg.CrossedLine && f.kept == nil
	if now {
		f.kept = req.Msg.Warning
	}
	return connect.NewResponse(&moneyrecapv1.EvaluateWarningResponse{Fired: f.kept != nil, FiredNow: now, Warning: f.kept}), nil
}

func (f *fakeNotify) RegisterDevice(_ context.Context, req *connect.Request[moneyrecapv1.RegisterDeviceRequest]) (*connect.Response[moneyrecapv1.RegisterDeviceResponse], error) {
	f.device = req.Msg
	return connect.NewResponse(&moneyrecapv1.RegisterDeviceResponse{}), nil
}
