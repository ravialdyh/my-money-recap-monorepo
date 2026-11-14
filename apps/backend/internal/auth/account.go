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

type ServiceAccount struct {
	Realm, Client, Secret string

	mu      sync.Mutex
	token   string
	expires time.Time
	now     func() time.Time
}

var realmClient = &http.Client{Timeout: 5 * time.Second}

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
