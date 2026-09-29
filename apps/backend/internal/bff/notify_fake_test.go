package bff

import (
	"context"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// fakeNotify is notify as the BFF sees it: kept is the stamped copy, which EvaluateWarning stamps on a crossed_line;
// prefs the preferences SavePrefs replaces unless saveErr; hang never answers, like a stopped container.
type fakeNotify struct {
	hang      bool
	kept      *moneyrecapv1.TargetWarning
	evaluated *moneyrecapv1.EvaluateWarningRequest
	device    *moneyrecapv1.RegisterDeviceRequest
	reads     int
	prefs     *moneyrecapv1.NotificationPrefs
	saved     *moneyrecapv1.SavePrefsRequest
	saveErr   error
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

func (f *fakeNotify) GetPrefs(ctx context.Context, _ *connect.Request[moneyrecapv1.GetPrefsRequest]) (*connect.Response[moneyrecapv1.GetPrefsResponse], error) {
	if f.hang {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return connect.NewResponse(&moneyrecapv1.GetPrefsResponse{Prefs: f.prefs}), nil
}

func (f *fakeNotify) SavePrefs(_ context.Context, req *connect.Request[moneyrecapv1.SavePrefsRequest]) (*connect.Response[moneyrecapv1.SavePrefsResponse], error) {
	f.saved = req.Msg
	if f.saveErr != nil {
		return nil, f.saveErr
	}
	f.prefs = req.Msg.Prefs
	return connect.NewResponse(&moneyrecapv1.SavePrefsResponse{Prefs: f.prefs}), nil
}
