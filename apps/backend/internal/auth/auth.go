// Package auth verifies the Keycloak JWT on every RPC (ARCHITECTURE §4.2, PRD §9).
package auth

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"
)

type ctxKey struct{}

// Claims is what a verified token tells the handler: the subject, the standard OIDC email
// claim, and the raw token so the BFF can forward it unchanged (§4.2 point 3).
type Claims struct{ Subject, Email, Raw string }

// Verifier checks issuer, audience, expiry and signature against the realm JWKS.
// The key set is fetched lazily and refreshed on an unknown key id.
type Verifier struct{ v *oidc.IDTokenVerifier }

func NewVerifier(ctx context.Context, issuer, jwksURL, audience string) *Verifier {
	return &Verifier{oidc.NewVerifier(issuer, oidc.NewRemoteKeySet(ctx, jwksURL), &oidc.Config{ClientID: audience})}
}

// Interceptor rejects any request without a valid bearer token and stores the
// token's Claims in the context for From.
func (v *Verifier) Interceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			raw, ok := strings.CutPrefix(req.Header().Get("Authorization"), "Bearer ")
			if !ok || raw == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing bearer token"))
			}
			tok, err := v.v.Verify(ctx, raw)
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
	c, _ := ctx.Value(ctxKey{}).(Claims)
	return c
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
