package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

const issuer = "http://localhost:8080/realms/mmr"

func TestInterceptor(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { json.NewEncoder(w).Encode(jwks) }))
	defer srv.Close()
	mint := func(aud string, exp time.Time) string {
		return testauth.Mint(t, key, "k1", jwt.Claims{Issuer: issuer, Subject: "sub-1", Audience: jwt.Audience{aud}, Expiry: jwt.NewNumericDate(exp), IssuedAt: jwt.NewNumericDate(time.Now())})
	}
	v := NewVerifier(context.Background(), issuer, srv.URL, "bff")
	var got string
	next := func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		got = From(ctx).Subject
		return nil, nil
	}
	call := func(header string) connect.Code {
		req := connect.NewRequest(&struct{}{})
		if header != "" {
			req.Header().Set("Authorization", header)
		}
		_, err := v.Interceptor()(next)(context.Background(), req)
		if err == nil {
			return 0
		}
		return connect.CodeOf(err)
	}
	in := time.Now().Add(time.Minute)
	cases := map[string]struct {
		header string
		want   connect.Code
	}{
		"valid":          {"Bearer " + mint("bff", in), 0},
		"missing":        {"", connect.CodeUnauthenticated},
		"not bearer":     {"Basic abc", connect.CodeUnauthenticated},
		"expired":        {"Bearer " + mint("bff", time.Now().Add(-time.Minute)), connect.CodeUnauthenticated},
		"wrong audience": {"Bearer " + mint("account", in), connect.CodeUnauthenticated},
		"garbage":        {"Bearer nope", connect.CodeUnauthenticated},
	}
	for name, c := range cases {
		got = ""
		if code := call(c.header); code != c.want {
			t.Errorf("%s: code %v, want %v", name, code, c.want)
		}
		if c.want == 0 && got != "sub-1" {
			t.Errorf("%s: subject %q", name, got)
		}
	}
}
