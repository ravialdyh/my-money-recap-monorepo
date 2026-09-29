package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
)

// ServiceAccount is a background job's own Keycloak client (ARCHITECTURE §4.2): Realm is the realm's
// address, parent of its token endpoint. Token caches the client-credentials token until a minute before it expires.
type ServiceAccount struct {
	Realm, Client, Secret string

	mu      sync.Mutex
	token   string
	expires time.Time
	now     func() time.Time
}

// realmClient bounds one token request.
var realmClient = &http.Client{Timeout: 5 * time.Second}

// Token is the account's access token, fetched with the client-credentials grant when none is cached or the
// cached one is within a minute of expiring. The error names the status alone, never the realm's body.
func (a *ServiceAccount) Token(ctx context.Context) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now
	if a.now != nil {
		now = a.now
	}
	if a.token != "" && now().Before(a.expires) {
		return a.token, nil
	}
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {a.Client}, "client_secret": {a.Secret}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.Realm+"/protocol/openid-connect/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := realmClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the %s token request answered %s", a.Client, res.Status)
	}
	var answer struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&answer); err != nil {
		return "", err
	}
	if answer.AccessToken == "" {
		return "", errors.New("the realm issued no service-account token")
	}
	a.token, a.expires = answer.AccessToken, now().Add(time.Duration(answer.ExpiresIn)*time.Second-time.Minute)
	return a.token, nil
}

// Bearer is a client interceptor that sends the account's token on every call; a token it cannot fetch is
// unavailable, and the call is not sent.
func (a *ServiceAccount) Bearer() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			tok, err := a.Token(ctx)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnavailable, err)
			}
			req.Header().Set("Authorization", "Bearer "+tok)
			return next(ctx, req)
		}
	}
}
