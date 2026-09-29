// Package auth verifies the Keycloak JWT on every RPC but the pre-authentication procedures, which
// are sent before a Member has a token (ARCHITECTURE §4.2, PRD §9), and fetches a background job's own token.
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

// Claims is what a verified token tells the handler: the subject, the standard OIDC email and
// given_name claims, the raw token so the BFF can forward it unchanged (§4.2 point 3), and Job for a
// background job's token, which owns no Member.
type Claims struct {
	Subject, Email, GivenName, Raw string
	Job                            bool
}

// Verifier checks issuer, audience, expiry and signature against the realm JWKS. Keys are cached
// and refetched on any miss, so while the realm is down a token with a cached key still verifies and
// any other is unavailable, never accepted.
type Verifier struct {
	v        *oidc.IDTokenVerifier
	audience string
}

func NewVerifier(ctx context.Context, issuer, jwksURL, audience string) *Verifier {
	keys := oidc.NewRemoteKeySet(oidc.ClientContext(ctx, jwksClient), jwksURL)
	return &Verifier{oidc.NewVerifier(issuer, keySet{keys}, &oidc.Config{SkipClientIDCheck: true}), audience}
}

// jwksClient gives a key fetch five seconds and fails, in its transport, any answer but a whole 200 of
// JSON, a redirect included, so every failed fetch (refused, DNS, timeout, TLS, 5xx, truncated body,
// a body that is not JSON) surfaces as a *url.Error.
var jwksClient = &http.Client{Timeout: 5 * time.Second, Transport: okOnly{http.DefaultTransport}}

// okOnly turns any answer but a whole 200 of JSON into a transport error; go-oidc would report it as
// plain text.
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

// keySet is the realm's RemoteKeySet. A fetch that fails, or that the caller's deadline cuts short, sets
// the *bool the interceptor put in the context, because go-oidc's Verify flattens that error into text.
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

// preAuth are the procedures sent before a Member has a token, which Interceptor calls through
// unverified. Membership is by procedure name, never by service, so the rest of the same service
// stays behind the check (ARCHITECTURE §4.2 point 3).
var preAuth = map[string]bool{
	moneyrecapv1connect.BffServiceRequestMagicLinkProcedure:  true,
	moneyrecapv1connect.NotifyServiceEmailMagicLinkProcedure: true,
}

// job is a background job's service account (ARCHITECTURE §4.2): its tokens name audience, and pass
// the check only on calls.
type job struct {
	audience string
	calls    map[string]bool
}

// jobs are the background jobs by the client id their tokens carry as azp. A job has no Member token,
// so it sends its own, which a service accepts only for the calls named here; any other is permission_denied.
var jobs = map[string]job{
	"notify": {audience: "ledger", calls: map[string]bool{
		moneyrecapv1connect.LedgerServiceGetSpentTodayProcedure:      true,
		moneyrecapv1connect.LedgerServiceListMonthsWithDataProcedure: true,
	}},
}

// Interceptor rejects any request without a valid bearer token, a pre-authentication procedure
// apart, and stores its Claims for From. A token it cannot check because the realm's keys are out
// of reach is unavailable, so the phone keeps its session (PRD D61).
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

// notValid logs why a token failed and answers a generic unauthenticated: go-oidc's text can quote the
// realm's key answer, which the phone has no use for.
func notValid(err error) error {
	log.Printf("auth: %v", err)
	return connect.NewError(connect.CodeUnauthenticated, errors.New("the bearer token is not valid"))
}

// NewContext stores verified Claims; the interceptor calls it, tests may too.
func NewContext(ctx context.Context, c Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// From is the verified token's Claims, zero outside the interceptor.
func From(ctx context.Context) Claims {
	claims, ok := ctx.Value(ctxKey{}).(Claims)
	if !ok {
		return Claims{}
	}
	return claims
}

// Forward is a client interceptor that sends the caller's own bearer token on to a service,
// unchanged, so the service verifies the same JWT again (§4.2 point 3).
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
