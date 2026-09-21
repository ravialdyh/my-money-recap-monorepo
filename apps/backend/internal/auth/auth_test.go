package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
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
		subject, code, _ := call(context.Background(), v, c.header)
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
	if _, got, _ := call(bg, cached, foreign); got != connect.CodeUnauthenticated {
		t.Errorf("key the reachable realm does not hold: %v, want unauthenticated", got)
	}
	if _, got, _ := call(bg, cached, valid); got != 0 {
		t.Fatalf("valid token: %v", got)
	}
	moved := httptest.NewServer(http.RedirectHandler(realm.URL, http.StatusFound))
	defer moved.Close()
	if _, got, _ := call(bg, NewVerifier(bg, issuer, moved.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("key URL redirecting to the realm: %v, want unavailable", got)
	}
	realm.Close()
	if _, got, _ := call(bg, cached, valid); got != 0 {
		t.Errorf("cached key while the realm is refused: %v, want accepted", got)
	}
	if _, got, _ := call(bg, cached, foreign); got != connect.CodeUnavailable {
		t.Errorf("uncached key while the realm is refused: %v, want unavailable", got)
	}
	if _, got, _ := call(bg, NewVerifier(bg, issuer, down.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm answering 503: %v, want unavailable", got)
	}
	if _, got, _ := call(bg, NewVerifier(bg, issuer, cut.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm cutting its 200 answer short: %v, want unavailable", got)
	}
	deadline, cancel := context.WithTimeout(bg, 100*time.Millisecond)
	defer cancel()
	if _, got, _ := call(deadline, NewVerifier(bg, issuer, hanging.URL, "bff"), valid); got != connect.CodeUnavailable {
		t.Errorf("realm hanging past the caller's deadline: %v, want unavailable", got)
	}
}

func TestInterceptorSendsNoRealmText(t *testing.T) {
	key := newKey(t)
	valid := "Bearer " + testauth.Mint(t, key, "k1", jwt.Claims{Issuer: issuer, Subject: "sub-1", Audience: jwt.Audience{"bff"}, Expiry: jwt.NewNumericDate(time.Now().Add(time.Minute))})
	cases := map[string]struct {
		contentType, body string
		want              connect.Code
	}{
		"200 that is not JSON":         {"text/html", "<html>maintenance at 10.0.0.7</html>", connect.CodeUnavailable},
		"200 of JSON that is not keys": {"application/json", `["10.0.0.7"]`, connect.CodeUnauthenticated},
	}
	for name, c := range cases {
		realm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", c.contentType)
			if _, err := io.WriteString(w, c.body); err != nil {
				t.Error(err)
			}
		}))
		t.Cleanup(realm.Close)
		_, code, text := call(context.Background(), NewVerifier(context.Background(), issuer, realm.URL, "bff"), valid)
		if code != c.want || strings.Contains(text, "10.0.0.7") {
			t.Errorf("%s: %v %q, want %v without the realm's text", name, code, text, c.want)
		}
	}
}

// The BFF reaches notify on a request that carries no token, because an exempted hop never filled
// the context Forward reads (ARCHITECTURE §4.2 point 3). Its siblings on the same service must
// still be refused, or the exemption is by service.
func TestAPreAuthProcedureIsTheOnlyOneCalledThroughWithoutAToken(t *testing.T) {
	v := NewVerifier(context.Background(), issuer, serveKey(t, newKey(t)).URL, "bff")
	srv := httptest.NewServer(mux(moneyrecapv1connect.NewNotifyServiceHandler(reached{}, connect.WithInterceptors(v.Interceptor()))))
	defer srv.Close()
	ctx := context.Background()

	for name, client := range map[string]moneyrecapv1connect.NotifyServiceClient{
		"no token":      moneyrecapv1connect.NewNotifyServiceClient(http.DefaultClient, srv.URL),
		"garbage token": moneyrecapv1connect.NewNotifyServiceClient(http.DefaultClient, srv.URL, testauth.Bearer("not.a.jwt")),
	} {
		if _, err := client.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{IdempotencyKey: "k"})); err != nil {
			t.Errorf("EmailMagicLink with %s: %v, want the handler reached", name, err)
		}
		_, err := client.GetPrefs(ctx, connect.NewRequest(&moneyrecapv1.GetPrefsRequest{}))
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("GetPrefs with %s: %v, want unauthenticated", name, err)
		}
	}
}

// The BFF's RequestMagicLink is outside the check one hop earlier, and for the same reason: it is
// sent before a Member has a token (#69, §4.2 point 2). Its siblings stay behind the check.
func TestTheBffsPreAuthProcedureIsTheOnlyOneCalledThroughWithoutAToken(t *testing.T) {
	v := NewVerifier(context.Background(), issuer, serveKey(t, newKey(t)).URL, "bff")
	srv := httptest.NewServer(mux(moneyrecapv1connect.NewBffServiceHandler(reachedBff{}, connect.WithInterceptors(v.Interceptor()))))
	defer srv.Close()
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, srv.URL)
	ctx := context.Background()

	if _, err := client.RequestMagicLink(ctx, connect.NewRequest(&moneyrecapv1.BffServiceRequestMagicLinkRequest{Email: "maya@example.test"})); err != nil {
		t.Errorf("RequestMagicLink with no token: %v, want the handler reached", err)
	}
	_, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("GetHome with no token: %v, want unauthenticated", err)
	}
}

// Forward sends the caller's own token on, and nothing at all when the context holds no verified
// claims, which is what an exempted hop leaves behind: the reason notify needs its own exemption
// rather than a token from the BFF (§4.2 point 3).
func TestForwardSendsAHeaderOnlyWhenTheContextHoldsAToken(t *testing.T) {
	for name, c := range map[string]struct {
		ctx  context.Context
		want string
	}{
		"verified claims": {NewContext(context.Background(), Claims{Subject: "sub-1", Raw: "tok"}), "Bearer tok"},
		"an exempted hop": {context.Background(), ""},
	} {
		req := connect.NewRequest(&struct{}{})
		if _, err := Forward()(func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
			return nil, nil
		})(c.ctx, req); err != nil {
			t.Fatal(err)
		}
		if got := req.Header().Get("Authorization"); got != c.want {
			t.Errorf("%s: sent %q, want %q", name, got, c.want)
		}
	}
}

// reached answers both procedures the exemption test calls, so a handler that ran is a nil error.
type reached struct {
	moneyrecapv1connect.UnimplementedNotifyServiceHandler
}

func (reached) EmailMagicLink(context.Context, *connect.Request[moneyrecapv1.EmailMagicLinkRequest]) (*connect.Response[moneyrecapv1.EmailMagicLinkResponse], error) {
	return connect.NewResponse(&moneyrecapv1.EmailMagicLinkResponse{}), nil
}

func (reached) GetPrefs(context.Context, *connect.Request[moneyrecapv1.GetPrefsRequest]) (*connect.Response[moneyrecapv1.GetPrefsResponse], error) {
	return connect.NewResponse(&moneyrecapv1.GetPrefsResponse{}), nil
}

// reachedBff is the same for the BFF's own pair: the exempted procedure and one behind the check.
type reachedBff struct {
	moneyrecapv1connect.UnimplementedBffServiceHandler
}

func (reachedBff) RequestMagicLink(context.Context, *connect.Request[moneyrecapv1.BffServiceRequestMagicLinkRequest]) (*connect.Response[moneyrecapv1.BffServiceRequestMagicLinkResponse], error) {
	return connect.NewResponse(&moneyrecapv1.BffServiceRequestMagicLinkResponse{}), nil
}

func (reachedBff) GetHome(context.Context, *connect.Request[moneyrecapv1.GetHomeRequest]) (*connect.Response[moneyrecapv1.GetHomeResponse], error) {
	return connect.NewResponse(&moneyrecapv1.GetHomeResponse{}), nil
}

func mux(path string, handler http.Handler) http.Handler {
	m := http.NewServeMux()
	m.Handle(path, handler)
	return m
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
// is 0 when the handler ran, subject is what the handler saw, and text is the error sent back.
func call(ctx context.Context, v *Verifier, header string) (subject string, code connect.Code, text string) {
	req := connect.NewRequest(&struct{}{})
	if header != "" {
		req.Header().Set("Authorization", header)
	}
	_, err := v.Interceptor()(func(ctx context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		subject = From(ctx).Subject
		return nil, nil
	})(ctx, req)
	if err != nil {
		return subject, connect.CodeOf(err), err.Error()
	}
	return subject, 0, ""
}
