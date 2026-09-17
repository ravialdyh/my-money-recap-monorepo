// Package testauth is what the tests against the Compose stack share: a token from the realm, its
// claims, a bearer client option, and a JWT signed by a key the realm never issued.
package testauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// Token posts form to the realm's token endpoint and returns the access token.
func Token(t *testing.T, keycloak string, form url.Values) string {
	t.Helper()
	res, err := http.PostForm(keycloak+"/realms/mmr/protocol/openid-connect/token", form)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil || body.AccessToken == "" {
		t.Fatalf("token request: status %d, err %v", res.StatusCode, err)
	}
	return body.AccessToken
}

// Claims are the payload fields the tests read, decoded without verification.
type Claims struct{ Iss, Sub, Email string }

func Decode(t *testing.T, tok string) Claims {
	t.Helper()
	raw, err := base64.RawURLEncoding.DecodeString(strings.Split(tok, ".")[1])
	if err != nil {
		t.Fatal(err)
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

// Bearer sends tok on every request; an empty tok sends no header.
func Bearer(tok string) connect.ClientOption {
	return connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if tok != "" {
				req.Header().Set("Authorization", "Bearer "+tok)
			}
			return next(ctx, req)
		}
	}))
}

// Mint signs claims with key under kid, RS256: a well-formed token from a key no JWKS holds.
func Mint(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.Claims) string {
	t.Helper()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", kid))
	if err != nil {
		t.Fatal(err)
	}
	s, err := jwt.Signed(signer).Claims(claims).Serialize()
	if err != nil {
		t.Fatal(err)
	}
	return s
}
