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

// Mailer is the slice of moneyrecapv1connect.NotifyServiceClient that RequestMagicLink needs (§7.3).
// notify is real for this one RPC since #68, while Notify stays a stub until slice 7.
type Mailer interface {
	EmailMagicLink(ctx context.Context, req *connect.Request[moneyrecapv1.EmailMagicLinkRequest]) (*connect.Response[moneyrecapv1.EmailMagicLinkResponse], error)
}

// Limiter is the slice of Redis the magic-link limit needs: one counter per key, expiring after
// window, as §4.4 counts with INCR and EXPIRE (A6).
type Limiter interface {
	Count(ctx context.Context, key string, window time.Duration) (int64, error)
}

// RedisLimiter is the BFF's first Redis client (§4.4, A6). TxPipeline makes the INCR and the EXPIRE
// one round trip, so a key can never outlive its window because the second call was lost.
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

// MagicLink is what RequestMagicLink needs of the realm, of Redis and of notify (A10, ARCHITECTURE §7.5).
type MagicLink struct {
	// Realm is Keycloak's realm from this process, parent of the token and login-token endpoints.
	Realm string
	// ClientSecret authenticates the bff service account, which holds manage-users.
	ClientSecret string
	// LinkBase is the configured address the link is built on; the login token is its lt query.
	LinkBase string
	// Limit counts §4.4's fifth Redis job, five requests per address per hour (PRD §11, A6).
	Limit Limiter
	// Mail is notify, reached on a request carrying no token (§4.2 point 3, #68).
	Mail Mailer
}

// linkLifetime is the 900 seconds A10 asks the extension for, which is the 15 minutes Check your
// inbox promises (PRD D32).
const linkLifetime = 900

// PRD §11's magic-link limit, a security requirement rather than fairness: five per address per
// fixed hour, counted in Redis because the edge cannot (§4.4, A6). limitTimeout bounds the count
// as realmClient bounds the realm: a closed door has to close fast.
const (
	limitPerWindow = 5
	limitWindow    = time.Hour
	limitTimeout   = 2 * time.Second
)

// realmClient bounds each of the two realm calls. The notify hop that follows carries notify's own
// ten-second cap on the SMTP conversation (internal/notify).
var realmClient = &http.Client{Timeout: 5 * time.Second}

// RequestMagicLink is the one BFF procedure outside the bearer check (internal/auth), so it reads
// the address from the request and never from the context's claims. It asks the realm for a login
// token and hands the address and the link to notify (A10, §7.5).
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

// withinLimit is PRD §11's five per address per hour, keyed by the address because the RPC carries
// no token and so has no subject. It fails closed, unlike §4.4's fail-open per-Member row: A6 sets
// the rule per job, and no edge limit can stand behind it.
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

// loginToken is the opaque lt:<uuid> the extension issues for addr, on A10's terms: 900 seconds,
// single use, the Member created if the address is new and the address marked verified (§7.5).
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

// serviceAccountToken is a client_credentials access token for the bff client, which is how the
// login-token endpoint sees the manage-users role. One per accepted request: five per address per
// hour is too few round trips to be worth a cache and its expiry.
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

// callRealm sends req and decodes a 200 of JSON into answer. Any other status is the error, named by
// path and status alone: the realm's body on this path can quote the address it was given.
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
