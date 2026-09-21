package bff

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// fakeMailer is notify: it keeps the last request and fails with sendErr when set.
type fakeMailer struct {
	got     *moneyrecapv1.EmailMagicLinkRequest
	calls   int
	sendErr error
}

func (f *fakeMailer) EmailMagicLink(_ context.Context, req *connect.Request[moneyrecapv1.EmailMagicLinkRequest]) (*connect.Response[moneyrecapv1.EmailMagicLinkResponse], error) {
	f.got, f.calls = req.Msg, f.calls+1
	if f.sendErr != nil {
		return nil, f.sendErr
	}
	return connect.NewResponse(&moneyrecapv1.EmailMagicLinkResponse{}), nil
}

// fakeRealm answers the two calls A10 makes, recording the login-token body; tokenStatus and
// loginStatus break either of them.
type fakeRealm struct {
	body                     map[string]any
	authorization            string
	tokenStatus, loginStatus int
	loginHint                string
}

func newRealm(t *testing.T, realm *fakeRealm) string {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /realms/mmr/protocol/openid-connect/token", func(w http.ResponseWriter, r *http.Request) {
		if realm.tokenStatus != 0 {
			w.WriteHeader(realm.tokenStatus)
			return
		}
		form, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(form), "grant_type=client_credentials") || !strings.Contains(string(form), "client_secret=s3cret") {
			t.Errorf("service-account request: %q", form)
		}
		_, _ = io.WriteString(w, `{"access_token":"service-account-token"}`)
	})
	mux.HandleFunc("POST /realms/mmr/login-token", func(w http.ResponseWriter, r *http.Request) {
		realm.authorization = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&realm.body)
		if realm.loginStatus != 0 {
			w.WriteHeader(realm.loginStatus)
			return
		}
		_, _ = io.WriteString(w, `{"login_hint":"`+realm.loginHint+`"}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL + "/realms/mmr"
}

func newMagicLinkService(t *testing.T, realm *fakeRealm, mailer Mailer) *Service {
	t.Helper()
	if realm.loginHint == "" {
		realm.loginHint = "lt:1e1a9b0c-0000-4000-8000-00000000abcd"
	}
	return &Service{MagicLink: MagicLink{
		Realm:        newRealm(t, realm),
		ClientSecret: "s3cret",
		LinkBase:     "http://localhost:8082/magic",
		Mail:         mailer,
	}}
}

func request(email, key string) *connect.Request[moneyrecapv1.BffServiceRequestMagicLinkRequest] {
	req := connect.NewRequest(&moneyrecapv1.BffServiceRequestMagicLinkRequest{Email: email})
	if key != "" {
		req.Header().Set("Idempotency-Key", key)
	}
	return req
}

// The handler reads the address from the request, never from the claims an exempted request never
// fills, and asks the realm on A10's exact terms (ARCHITECTURE §7.5).
func TestRequestMagicLinkAsksTheRealmOnA10sTermsAndMailsTheLink(t *testing.T) {
	realm, mailer := &fakeRealm{}, &fakeMailer{}
	s := newMagicLinkService(t, realm, mailer)
	claimed := auth.NewContext(context.Background(), auth.Claims{Subject: "sub-1", Email: "someone-else@example.test", Raw: "tok"})
	if _, err := s.RequestMagicLink(claimed, request("maya@example.test", "key-1")); err != nil {
		t.Fatal(err)
	}
	if realm.authorization != "Bearer service-account-token" {
		t.Errorf("login-token authorization %q, want the bff service account's token", realm.authorization)
	}
	if realm.body["email"] != "maya@example.test" || realm.body["client_id"] != "ios" ||
		realm.body["expiration_seconds"] != float64(linkLifetime) || realm.body["reusable"] != false ||
		realm.body["force_create"] != true || realm.body["set_email_verified"] != true {
		t.Errorf("login-token body %v", realm.body)
	}
	if mailer.got.Email != "maya@example.test" || mailer.got.IdempotencyKey != "key-1" ||
		mailer.got.Link != "http://localhost:8082/magic?lt=lt%3A1e1a9b0c-0000-4000-8000-00000000abcd" {
		t.Errorf("notify received %v", mailer.got)
	}
}

// Nothing sets Idempotency-Key until slice 4, and notify rejects an empty key, so an absent header
// mints a fresh UUID per accepted request rather than deduping every later send into silence (A14).
func TestRequestMagicLinkMintsAFreshKeyWhenTheHeaderIsAbsent(t *testing.T) {
	mailer := &fakeMailer{}
	s := newMagicLinkService(t, &fakeRealm{}, mailer)
	var keys []string
	for range 2 {
		if _, err := s.RequestMagicLink(context.Background(), request("maya@example.test", "")); err != nil {
			t.Fatal(err)
		}
		keys = append(keys, mailer.got.IdempotencyKey)
	}
	if keys[0] == keys[1] || len(keys[0]) != 36 || keys[0][14] != '4' {
		t.Errorf("minted %q then %q, want two different version 4 UUIDs", keys[0], keys[1])
	}
}

// A realm that answers nothing at all, Keycloak stopped, is unavailable too: the transport error
// comes back before any status does.
func TestRequestMagicLinkWhenTheRealmIsUnreachable(t *testing.T) {
	mailer := &fakeMailer{}
	s := newMagicLinkService(t, &fakeRealm{}, mailer)
	s.MagicLink.Realm = "http://127.0.0.1:1/realms/mmr"
	_, err := s.RequestMagicLink(context.Background(), request("maya@example.test", ""))
	if connect.CodeOf(err) != connect.CodeUnavailable || mailer.calls != 0 {
		t.Errorf("%v after %d sends, want unavailable after 0", err, mailer.calls)
	}
}

func TestRequestMagicLinkFailures(t *testing.T) {
	for name, c := range map[string]struct {
		realm  *fakeRealm
		mailer *fakeMailer
		email  string
		want   connect.Code
		mailed int
	}{
		"not an address":          {&fakeRealm{}, &fakeMailer{}, "not-an-address", connect.CodeInvalidArgument, 0},
		"empty address":           {&fakeRealm{}, &fakeMailer{}, "", connect.CodeInvalidArgument, 0},
		"no service account":      {&fakeRealm{tokenStatus: http.StatusUnauthorized}, &fakeMailer{}, "maya@example.test", connect.CodeUnavailable, 0},
		"login-token refuses":     {&fakeRealm{loginStatus: http.StatusBadRequest}, &fakeMailer{}, "maya@example.test", connect.CodeUnavailable, 0},
		"login-token answers odd": {&fakeRealm{loginHint: "no"}, &fakeMailer{}, "maya@example.test", connect.CodeUnavailable, 0},
		"notify refuses":          {&fakeRealm{}, &fakeMailer{sendErr: connect.NewError(connect.CodeInvalidArgument, io.EOF)}, "maya@example.test", connect.CodeInvalidArgument, 1},
	} {
		s := newMagicLinkService(t, c.realm, c.mailer)
		_, err := s.RequestMagicLink(context.Background(), request(c.email, ""))
		if connect.CodeOf(err) != c.want || c.mailer.calls != c.mailed {
			t.Errorf("%s: %v after %d sends, want %v after %d", name, err, c.mailer.calls, c.want, c.mailed)
		}
	}
}
