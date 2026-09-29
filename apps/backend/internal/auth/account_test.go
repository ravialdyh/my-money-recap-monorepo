package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
)

// A job's token is fetched with the client-credentials grant, reused until a minute before it expires, then fetched
// again; a refusal names the status alone, and Bearer sends the token or, failing it, nothing at all (§4.2).
func TestAServiceAccountCachesItsTokenUntilItExpires(t *testing.T) {
	fetches, refuse := 0, false
	realm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil || r.URL.Path != "/protocol/openid-connect/token" || r.Form.Get("grant_type") != "client_credentials" ||
			r.Form.Get("client_id") != "notify" || r.Form.Get("client_secret") != "s3cret" {
			t.Errorf("token request %s %v", r.URL.Path, r.Form)
		}
		if refuse {
			http.Error(w, `{"error":"unauthorized_client","secret":"s3cret"}`, http.StatusUnauthorized)
			return
		}
		fetches++
		fmt.Fprintf(w, `{"access_token":"tok-%d","expires_in":300}`, fetches)
	}))
	defer realm.Close()
	clock := time.Now()
	a := &ServiceAccount{Realm: realm.URL, Client: "notify", Secret: "s3cret", now: func() time.Time { return clock }}
	ctx := context.Background()
	for _, step := range []struct {
		after time.Duration
		want  string
	}{{0, "tok-1"}, {3 * time.Minute, "tok-1"}, {time.Minute + time.Second, "tok-2"}} {
		clock = clock.Add(step.after)
		if got, err := a.Token(ctx); err != nil || got != step.want {
			t.Errorf("after %v more: %q, %v, want %q", step.after, got, err, step.want)
		}
	}

	var sent string
	send := a.Bearer()(func(_ context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		sent = req.Header().Get("Authorization")
		return nil, nil
	})
	if _, err := send(ctx, connect.NewRequest(&struct{}{})); err != nil || sent != "Bearer tok-2" {
		t.Errorf("Bearer sent %q, %v, want the cached token", sent, err)
	}
	clock, refuse, sent = clock.Add(time.Hour), true, ""
	_, err := send(ctx, connect.NewRequest(&struct{}{}))
	if connect.CodeOf(err) != connect.CodeUnavailable || sent != "" || strings.Contains(err.Error(), "s3cret") || !strings.Contains(err.Error(), "401") {
		t.Errorf("a refused token request: %v, sent %q; want unavailable, naming the status alone, with nothing sent", err, sent)
	}
}
