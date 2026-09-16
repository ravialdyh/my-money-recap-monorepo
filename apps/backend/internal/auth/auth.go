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

// Verifier checks issuer, audience, expiry and signature against the realm JWKS.
// The key set is fetched lazily and refreshed on an unknown key id.
type Verifier struct{ v *oidc.IDTokenVerifier }

func NewVerifier(ctx context.Context, issuer, jwksURL, audience string) *Verifier {
	return &Verifier{oidc.NewVerifier(issuer, oidc.NewRemoteKeySet(ctx, jwksURL), &oidc.Config{ClientID: audience})}
}

// Interceptor rejects any request without a valid bearer token and stores the
// token's subject in the context for Subject.
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
			return next(context.WithValue(ctx, ctxKey{}, tok.Subject), req)
		}
	}
}

// Subject is the verified token's subject claim, empty outside the interceptor.
func Subject(ctx context.Context) string {
	s, _ := ctx.Value(ctxKey{}).(string)
	return s
}
