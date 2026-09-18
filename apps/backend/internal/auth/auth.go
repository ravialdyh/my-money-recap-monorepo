// Package auth verifies the Keycloak JWT on every RPC (ARCHITECTURE §4.2, PRD §9).
package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"
)

type ctxKey struct{}

type fetchFailedKey struct{}

// Claims is what a verified token tells the handler: the subject, the standard OIDC email
// claim, and the raw token so the BFF can forward it unchanged (§4.2 point 3).
type Claims struct{ Subject, Email, Raw string }

// Verifier checks issuer, audience, expiry and signature against the realm JWKS. Keys are cached
// and refetched on any miss, so while the realm is down a token with a cached key still verifies and
// any other is unavailable, never accepted.
type Verifier struct{ v *oidc.IDTokenVerifier }

func NewVerifier(ctx context.Context, issuer, jwksURL, audience string) *Verifier {
	keys := oidc.NewRemoteKeySet(oidc.ClientContext(ctx, jwksClient), jwksURL)
	return &Verifier{oidc.NewVerifier(issuer, keySet{keys}, &oidc.Config{ClientID: audience})}
}

// jwksClient gives a key fetch five seconds and fails a non-200 answer in its transport, so every
// failed fetch (refused, DNS, timeout, TLS, 5xx) surfaces as a *url.Error.
var jwksClient = &http.Client{Timeout: 5 * time.Second, Transport: okOnly{http.DefaultTransport}}

// okOnly turns any answer but 200 into a transport error; go-oidc would report it as plain text.
type okOnly struct{ next http.RoundTripper }

func (t okOnly) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("realm keys answered %s", res.Status)
	}
	return res, nil
}

// keySet is the realm's RemoteKeySet. A failed fetch sets the *bool the interceptor put in the
// context, because go-oidc's Verify flattens that error into text.
type keySet struct{ *oidc.RemoteKeySet }

func (k keySet) VerifySignature(ctx context.Context, jwt string) ([]byte, error) {
	payload, err := k.RemoteKeySet.VerifySignature(ctx, jwt)
	var fetch *url.Error
	if failed, ok := ctx.Value(fetchFailedKey{}).(*bool); ok && errors.As(err, &fetch) {
		*failed = true
	}
	return payload, err
}

// Interceptor rejects any request without a valid bearer token and stores its Claims in the context
// for From. A token it cannot check because the realm's keys are out of reach is unavailable, so the
// phone keeps its session (PRD D61).
func (v *Verifier) Interceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
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
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			var extra struct {
				Email string `json:"email"`
			}
			if err := tok.Claims(&extra); err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}
			return next(NewContext(ctx, Claims{Subject: tok.Subject, Email: extra.Email, Raw: raw}), req)
		}
	}
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
