package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
)

type ctxKey struct{}

type fetchFailedKey struct{}

type Claims struct {
	Subject, Email, GivenName, Raw string
	Job                            bool
}

type Verifier struct {
	v        *oidc.IDTokenVerifier
	audience string
}

func NewVerifier(ctx context.Context, issuer, jwksURL, audience string) *Verifier {
	keys := oidc.NewRemoteKeySet(oidc.ClientContext(ctx, jwksClient), jwksURL)
	return &Verifier{oidc.NewVerifier(issuer, keySet{keys}, &oidc.Config{SkipClientIDCheck: true}), audience}
}

var jwksClient = &http.Client{Timeout: 5 * time.Second, Transport: okOnly{http.DefaultTransport}}

type okOnly struct{ next http.RoundTripper }

func (t okOnly) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("realm keys answered %s", res.Status)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if !json.Valid(body) {
		return nil, errors.New("realm keys answered 200 with a body that is not JSON")
	}
	res.Body = io.NopCloser(bytes.NewReader(body))
	return res, nil
}

type keySet struct{ *oidc.RemoteKeySet }

func (k keySet) VerifySignature(ctx context.Context, jwt string) ([]byte, error) {
	payload, err := k.RemoteKeySet.VerifySignature(ctx, jwt)
	var fetch *url.Error
	failed, ok := ctx.Value(fetchFailedKey{}).(*bool)
	if ok && err != nil && (errors.As(err, &fetch) || errors.Is(err, ctx.Err())) {
		*failed = true
	}
	return payload, err
}

var preAuth = map[string]bool{
	moneyrecapv1connect.BffServiceRequestMagicLinkProcedure:  true,
	moneyrecapv1connect.NotifyServiceEmailMagicLinkProcedure: true,
}

type job struct {
	audience string
	calls    map[string]bool
}

var jobs = map[string]job{
	"notify": {audience: "ledger", calls: map[string]bool{
		moneyrecapv1connect.LedgerServiceGetSpentTodayProcedure:      true,
		moneyrecapv1connect.LedgerServiceListMonthsWithDataProcedure: true,
	}},
}

func (v *Verifier) Interceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if preAuth[req.Spec().Procedure] {
				return next(ctx, req)
			}
			raw, ok := strings.CutPrefix(req.Header().Get("Authorization"), "Bearer ")
			if !ok || raw == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing bearer token"))
			}
			fetchFailed := false
			tok, err := v.v.Verify(context.WithValue(ctx, fetchFailedKey{}, &fetchFailed), raw)
			if err != nil && fetchFailed {
				log.Printf("auth: %v", err)
				return nil, connect.NewError(connect.CodeUnavailable, errors.New("the realm's signing keys are out of reach"))
			}
			if err != nil {
				return nil, notValid(err)
			}
			var extra struct {
				Email     string `json:"email"`
				GivenName string `json:"given_name"`
				Azp       string `json:"azp"`
			}
			if err := tok.Claims(&extra); err != nil {
				return nil, notValid(err)
			}
			if j, ok := jobs[extra.Azp]; ok && slices.Contains(tok.Audience, j.audience) {
				if !j.calls[req.Spec().Procedure] {
					return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("the %s job may not call this", extra.Azp))
				}
				return next(NewContext(ctx, Claims{Subject: tok.Subject, Raw: raw, Job: true}), req)
			}
			if !slices.Contains(tok.Audience, v.audience) {
				return nil, notValid(fmt.Errorf("audience %v lacks %q", tok.Audience, v.audience))
			}
			return next(NewContext(ctx, Claims{Subject: tok.Subject, Email: extra.Email, GivenName: extra.GivenName, Raw: raw}), req)
		}
	}
}

func notValid(err error) error {
	log.Printf("auth: %v", err)
	return connect.NewError(connect.CodeUnauthenticated, errors.New("the bearer token is not valid"))
}

func NewContext(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

func From(ctx context.Context) Claims {
	claims, ok := ctx.Value(ctxKey{}).(Claims)
	if !ok {
		return Claims{}
	}
	return claims
}

func Forward() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if raw := From(ctx).Raw; raw != "" {
				req.Header().Set("Authorization", "Bearer "+raw)
			}
			return next(ctx, req)
		}
	}
}
