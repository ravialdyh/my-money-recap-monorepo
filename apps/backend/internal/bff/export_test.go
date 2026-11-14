package bff

import (
	"errors"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// RequestExport refuses a keyless call before ledger and notify, hands notify ledger's CSV under the key, and passes a
// refusal on with its EmailBounced detail, which is how the phone names the address (PRD §6.16).
func TestRequestExportHandsLedgersCSVToNotify(t *testing.T) {
	notify := &fakeNotify{}
	s := newService(t, &fakeMembers{profile: onboarded()}, notify)
	req := connect.NewRequest(&moneyrecapv1.BffServiceRequestExportRequest{})
	if _, err := s.RequestExport(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument || notify.exported != nil {
		t.Fatalf("no key: %v, notify reached %v; want invalid_argument and no call", err, notify.exported != nil)
	}
	req.Header().Set("Idempotency-Key", "key-1")
	if _, err := s.RequestExport(ctx, req); err != nil || notify.exported.GetIdempotencyKey() != "key-1" || notify.exported.GetMemberId() != "m-1" ||
		string(notify.exported.GetCsv()) != "date,category,amount,currency,note\r\n" {
		t.Fatalf("RequestExport: %v; notify got %v", err, notify.exported)
	}

	refused := connect.NewError(connect.CodeFailedPrecondition, errors.New("the mail host refused the address"))
	detail, err := connect.NewErrorDetail(&moneyrecapv1.EmailBounced{Address: "test@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	refused.AddDetail(detail)
	notify.exportErr = refused
	_, err = s.RequestExport(ctx, req)
	var ce *connect.Error
	if !errors.As(err, &ce) || ce.Code() != connect.CodeFailedPrecondition || len(ce.Details()) != 1 {
		t.Fatalf("a refused address: %v, want failed_precondition carrying EmailBounced", err)
	}
	v, err := ce.Details()[0].Value()
	if bounced, ok := v.(*moneyrecapv1.EmailBounced); err != nil || !ok || bounced.GetAddress() != "test@example.com" {
		t.Errorf("the detail %v (%v), want EmailBounced naming the address", v, err)
	}
	notify.exportErr = connect.NewError(connect.CodeUnavailable, errors.New("the export email was not sent"))
	if _, err := s.RequestExport(ctx, req); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("a failed send: %v, want unavailable", err)
	}
	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeUnavailable, errors.New("down"))})
	notify.exported = nil
	if _, err := s.RequestExport(ctx, req); connect.CodeOf(err) != connect.CodeUnavailable || notify.exported != nil {
		t.Errorf("ledger down: %v, notify reached %v; want unavailable and no email", err, notify.exported != nil)
	}
}
