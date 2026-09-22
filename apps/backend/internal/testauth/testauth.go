// Package testauth is what the tests against the Compose stack share: a token from the realm, its
// claims, a bearer client option, a JWT signed by a key the realm never issued, and Mailpit's inbox.
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
	"time"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// Token posts form to the realm's token endpoint and returns the access token, failing the test on a
// refusal.
func Token(t *testing.T, keycloak string, form url.Values) string {
	t.Helper()
	access, refusal := Exchange(t, keycloak, form)
	if access == "" {
		t.Fatalf("token request: %s", refusal)
	}
	return access
}

// UserToken is the access token of the realm's one test user, `test`, by the password grant on the
// ios client.
func UserToken(t *testing.T, keycloak, password string) string {
	t.Helper()
	return Token(t, keycloak, url.Values{"grant_type": {"password"}, "client_id": {"ios"}, "username": {"test"}, "password": {password}})
}

// Exchange posts form to the realm's token endpoint and answers the access token, or "" and the
// realm's refusal as its status and error_description.
func Exchange(t *testing.T, keycloak string, form url.Values) (access, refusal string) {
	t.Helper()
	endpoint := keycloak + "/realms/mmr/protocol/openid-connect/token"
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		AccessToken      string `json:"access_token"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("token request: status %d, err %v", res.StatusCode, err)
	}
	return body.AccessToken, res.Status + " " + body.ErrorDescription
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

// Inbox is the ids of every message Mailpit holds for address, polled for three seconds until it
// holds at least want, so an assertion on two waits for the second rather than reading the first poll.
func Inbox(t *testing.T, mailpit, address string, want int) []string {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		var found struct {
			Messages []struct{ ID string }
		}
		getJSON(t, mailpit+"/api/v1/search?query="+url.QueryEscape("to:"+address), &found)
		ids := make([]string, 0, len(found.Messages))
		for _, m := range found.Messages {
			ids = append(ids, m.ID)
		}
		if len(ids) >= want || time.Now().After(deadline) {
			return ids
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// MessageText is the plain-text body of the Mailpit message id.
func MessageText(t *testing.T, mailpit, id string) string {
	t.Helper()
	var message struct{ Text string }
	getJSON(t, mailpit+"/api/v1/message/"+id, &message)
	return message.Text
}

// getJSON decodes Mailpit's answer into out, failing the test on anything but a 200 of JSON.
func getJSON(t *testing.T, address string, out any) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, address, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("%s: status %d", address, res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		t.Fatal(err)
	}
}
