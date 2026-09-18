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
	key := newKey(t)
	mint := func(aud string, exp time.Time) string {
		return testauth.Mint(t, key, "k1", jwt.Claims{Issuer: issuer, Subject: "sub-1", Audience: jwt.Audience{aud}, Expiry: jwt.NewNumericDate(exp), IssuedAt: jwt.NewNumericDate(time.Now())})
	}
	v := NewVerifier(context.Background(), issuer, serveKey(t, key).URL, "bff")
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
		subject, code := call(context.Background(), v, c.header)
		if code != c.want {
			t.Errorf("%s: code %v, want %v", name, code, c.want)
		}
		if c.want == 0 && subject != "sub-1" {
			t.Errorf("%s: subject %q", name, subject)
		}
	}
}

func TestInterceptorKeyFetchFailureIsUnavailable(t *testing.T) {
	key, stranger := newKey(t), newKey(t)
	realm := serveKey(t, key)
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
	valid, foreign := "Bearer "+testauth.Mint(t, key, "k1", claims), "Bearer "+testauth.Mint(t, stranger, "k2", claims)
	bg := context.Background()

	cached := NewVerifier(bg, issuer, realm.URL, "bff")
	if _, got := call(bg, cached, foreign); got != connect.CodeUnauthenticated {
		t.Errorf("key the reachable realm does not hold: %v, want unauthenticated", got)
	}
	if _, got := call(bg, cached, valid); got != 0 {
		t.Fatalf("valid token: %v", got)
	}
	realm.Close()
	if _, got := call(bg, cached, valid); got != 0 {
		t.Errorf("cached key while the realm is refused: %v, want accepted", got)
	}
	if _, got := call(bg, cached, foreign); got != connect.CodeUnavailable {
		t.Errorf("uncached key while the realm is refused: %v, want unavailable", got)
	}
	if _, got := call(bg, NewVerifier(bg, issuer, down.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm answering 503: %v, want unavailable", got)
	}
	if _, got := call(bg, NewVerifier(bg, issuer, cut.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm cutting its 200 answer short: %v, want unavailable", got)
	}
	deadline, cancel := context.WithTimeout(bg, 100*time.Millisecond)
	defer cancel()
	if _, got := call(deadline, NewVerifier(bg, issuer, hanging.URL, "bff"), valid); got != connect.CodeUnavailable {
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

// serveKey is a realm whose JWKS holds key's public half under kid k1, closed when the test ends.
func serveKey(t *testing.T, key *rsa.PrivateKey) *httptest.Server {
	t.Helper()
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"}}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if err := json.NewEncoder(w).Encode(jwks); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// call sends one request through v's interceptor with header as Authorization (none when empty); code
// is 0 when the handler ran, and subject is what the handler saw.
func call(ctx context.Context, v *Verifier, header string) (subject string, code connect.Code) {
	req := connect.NewRequest(&struct{}{})
	if header != "" {
		req.Header().Set("Authorization", header)
	}
	_, err := v.Interceptor()(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		subject = From(ctx).Subject
		return nil, nil
	})(ctx, req)
	if err != nil {
		return subject, connect.CodeOf(err)
	}
	return subject, 0
}
