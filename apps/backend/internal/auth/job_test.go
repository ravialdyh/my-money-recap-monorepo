package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// A background job's token passes only on the calls named for its job, marked Job; elsewhere it is
// permission_denied, and a service account no job names is refused like any other audience (§4.2).
func TestAJobsTokenPassesOnlyItsOwnCalls(t *testing.T) {
	key := newKey(t)
	mint := func(azp, aud string) connect.ClientOption {
		claims := struct {
			jwt.Claims
			Azp string `json:"azp"`
		}{jwt.Claims{Issuer: issuer, Subject: "service-account", Audience: jwt.Audience{aud, "account"}, Expiry: jwt.NewNumericDate(time.Now().Add(time.Minute))}, azp}
		signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "k1"))
		if err != nil {
			t.Fatal(err)
		}
		tok, err := jwt.Signed(signer).Claims(claims).Serialize()
		if err != nil {
			t.Fatal(err)
		}
		return testauth.Bearer(tok)
	}
	v := NewVerifier(context.Background(), issuer, serveKey(t, key).URL, "bff")
	srv := httptest.NewServer(mux(moneyrecapv1connect.NewLedgerServiceHandler(reachedLedger{}, connect.WithInterceptors(v.Interceptor()))))
	defer srv.Close()
	ctx := context.Background()
	for name, c := range map[string]struct {
		token                connect.ClientOption
		spentToday, settings connect.Code
	}{
		"notify's token": {mint("notify", "ledger"), 0, connect.CodePermissionDenied},
		"notify's client without the ledger audience": {mint("notify", "notify"), connect.CodeUnauthenticated, connect.CodeUnauthenticated},
		"a service account no job names":              {mint("members", "ledger"), connect.CodeUnauthenticated, connect.CodeUnauthenticated},
		"a Member's token":                            {mint("ios", "bff"), connect.CodeFailedPrecondition, 0},
	} {
		client := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, srv.URL, c.token)
		_, err := client.GetSpentToday(ctx, connect.NewRequest(&moneyrecapv1.GetSpentTodayRequest{}))
		if got := connect.CodeOf(err); err != nil && got != c.spentToday || err == nil && c.spentToday != 0 {
			t.Errorf("%s, GetSpentToday: %v, want %v", name, err, c.spentToday)
		}
		_, err = client.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{}))
		if got := connect.CodeOf(err); err != nil && got != c.settings || err == nil && c.settings != 0 {
			t.Errorf("%s, GetSettings: %v, want %v", name, err, c.settings)
		}
	}
}

// reachedLedger answers GetSpentToday only for a job's claims, failed_precondition otherwise, so the test sees Job.
type reachedLedger struct {
	moneyrecapv1connect.UnimplementedLedgerServiceHandler
}

func (reachedLedger) GetSpentToday(ctx context.Context, _ *connect.Request[moneyrecapv1.GetSpentTodayRequest]) (*connect.Response[moneyrecapv1.GetSpentTodayResponse], error) {
	if !From(ctx).Job {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("not a job"))
	}
	return connect.NewResponse(&moneyrecapv1.GetSpentTodayResponse{}), nil
}

func (reachedLedger) GetSettings(context.Context, *connect.Request[moneyrecapv1.GetSettingsRequest]) (*connect.Response[moneyrecapv1.GetSettingsResponse], error) {
	return connect.NewResponse(&moneyrecapv1.GetSettingsResponse{}), nil
}
