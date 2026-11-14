package bff

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"uuid"

	"connectrpc.com/connect"
	"github.com/redis/go-redis/v9"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

type Mailer interface {
	EmailMagicLink(ctx context.Context, req *connect.Request[moneyrecapv1.EmailMagicLinkRequest]) (*connect.Response[moneyrecapv1.EmailMagicLinkResponse], error)
}

type Limiter interface {
	Count(ctx context.Context, key string, window time.Duration) (int64, error)
}

type RedisLimiter struct {
	Client redis.UniversalClient
}

func (r RedisLimiter) Count(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := r.Client.TxPipeline()
	count := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return count.Val(), nil
}

type MagicLink struct {
	Realm        string
	ClientSecret string
	LinkBase     string
	Limit        Limiter
	Mail         Mailer
}

const linkLifetime = 900

const (
	limitPerWindow = 5
	limitWindow    = time.Hour
	limitTimeout   = 2 * time.Second
)

var realmClient = &http.Client{Timeout: 5 * time.Second}

func (s *Service) RequestMagicLink(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceRequestMagicLinkRequest]) (*connect.Response[moneyrecapv1.BffServiceRequestMagicLinkResponse], error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(req.Msg.Email))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("email is not an address"))
	}
	if err := s.MagicLink.withinLimit(ctx, addr.Address); err != nil {
		return nil, err
	}
	key := req.Header().Get("Idempotency-Key")
	if key == "" {
		key = uuid.NewV4().String()
	}
	token, err := s.MagicLink.loginToken(ctx, addr.Address)
	if err != nil {
		return nil, requiredCallFailed("keycloak.login-token", connect.CodeUnavailable, err)
	}
	send := &moneyrecapv1.EmailMagicLinkRequest{
		IdempotencyKey: key,
		Email:          addr.Address,
		Link:           s.MagicLink.LinkBase + "?lt=" + url.QueryEscape(token),
	}
	if _, err := s.MagicLink.Mail.EmailMagicLink(ctx, connect.NewRequest(send)); err != nil {
		return nil, requiredCallFailed("notify.EmailMagicLink", connect.CodeOf(err), err)
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceRequestMagicLinkResponse{}), nil
}

func (m MagicLink) withinLimit(ctx context.Context, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, limitTimeout)
	defer cancel()
	window := time.Now().Unix() / int64(limitWindow/time.Second)
	count, err := m.Limit.Count(ctx, fmt.Sprintf("ml:%s:%d", strings.ToLower(addr), window), limitWindow)
	if err != nil {
		return requiredCallFailed("redis.magic-link-limit", connect.CodeResourceExhausted, err)
	}
	if count > limitPerWindow {
		return connect.NewError(connect.CodeResourceExhausted, errors.New("too many magic-link requests for this address"))
	}
	return nil
}

func (m MagicLink) loginToken(ctx context.Context, addr string) (string, error) {
	account, err := m.serviceAccountToken(ctx)
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(map[string]any{
		"email": addr, "client_id": "ios", "expiration_seconds": linkLifetime,
		"reusable": false, "force_create": true, "set_email_verified": true,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.Realm+"/login-token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+account)
	req.Header.Set("Content-Type", "application/json")
	var answer struct {
		LoginHint string `json:"login_hint"`
	}
	if err := callRealm(req, &answer); err != nil {
		return "", err
	}
	if !strings.HasPrefix(answer.LoginHint, "lt:") {
		return "", errors.New("the login-token endpoint answered without a login token")
	}
	return answer.LoginHint, nil
}

func (m MagicLink) serviceAccountToken(ctx context.Context) (string, error) {
	form := url.Values{"grant_type": {"client_credentials"}, "client_id": {"bff"}, "client_secret": {m.ClientSecret}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.Realm+"/protocol/openid-connect/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var answer struct {
		AccessToken string `json:"access_token"`
	}
	if err := callRealm(req, &answer); err != nil {
		return "", err
	}
	if answer.AccessToken == "" {
		return "", errors.New("the realm issued no service-account token")
	}
	return answer.AccessToken, nil
}

func callRealm(req *http.Request, answer any) error {
	res, err := realmClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", req.URL.Path, res.Status)
	}
	return json.NewDecoder(res.Body).Decode(answer)
}
