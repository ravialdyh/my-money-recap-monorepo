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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(jwks); err != nil {
			t.Error(err)
		}
	}))
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

func TestInterceptorKeyFetchFailureIsUnavailable(t *testing.T) {
	key, stranger := newKey(t), newKey(t)
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"}}}
	realm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(jwks); err != nil {
			t.Error(err)
		}
	}))
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer down.Close()
	cut := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "1000")
		if _, err := w.Write([]byte(`{"keys":[`)); err != nil {
			t.Error(err)
		}
	}))
	defer cut.Close()
	hung := make(chan struct{})
	hanging := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-hung }))
	defer hanging.Close()
	defer close(hung)
	claims := jwt.Claims{Issuer: issuer, Subject: "sub-1", Audience: jwt.Audience{"bff"}, Expiry: jwt.NewNumericDate(time.Now().Add(time.Minute))}
	valid, foreign := testauth.Mint(t, key, "k1", claims), testauth.Mint(t, stranger, "k2", claims)
	code := func(ctx context.Context, v *Verifier, tok string) connect.Code {
		req := connect.NewRequest(&struct{}{})
		req.Header().Set("Authorization", "Bearer "+tok)
		_, err := v.Interceptor()(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) { return nil, nil })(ctx, req)
		if err == nil {
			return 0
		}
		return connect.CodeOf(err)
	}
	bg := context.Background()

	cached := NewVerifier(bg, issuer, realm.URL, "bff")
	if got := code(bg, cached, foreign); got != connect.CodeUnauthenticated {
		t.Errorf("key the reachable realm does not hold: %v, want unauthenticated", got)
	}
	if got := code(bg, cached, valid); got != 0 {
		t.Fatalf("valid token: %v", got)
	}
	realm.Close()
	if got := code(bg, cached, valid); got != 0 {
		t.Errorf("cached key while the realm is refused: %v, want accepted", got)
	}
	if got := code(bg, cached, foreign); got != connect.CodeUnavailable {
		t.Errorf("uncached key while the realm is refused: %v, want unavailable", got)
	}
	if got := code(bg, NewVerifier(bg, issuer, down.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm answering 503: %v, want unavailable", got)
	}
	if got := code(bg, NewVerifier(bg, issuer, cut.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm cutting its 200 answer short: %v, want unavailable", got)
	}
	deadline, cancel := context.WithTimeout(bg, 100*time.Millisecond)
	defer cancel()
	if got := code(deadline, NewVerifier(bg, issuer, hanging.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm hanging past the caller's deadline: %v, want unavailable", got)
	}
}

func newKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key
}
