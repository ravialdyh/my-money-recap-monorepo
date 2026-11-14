package notify

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/jackc/pgx/v5"
)

// tokenLife is how long one signed Apple token is reused: Apple refuses a token older than an hour and one replaced
// more often than every 20 minutes.
const tokenLife = 40 * time.Minute

// pushTimeout bounds one push's whole background run, its database writes and every device's request.
const pushTimeout = 30 * time.Second

// Apple sends pushes through Apple's push service with token authentication (ARCHITECTURE §7.3): Host is its origin,
// Topic the app's bundle id, and a nil Key means no key is configured, so every push is recorded and not sent.
type Apple struct {
	Host, Topic, TeamID, KeyID string
	Key                        *ecdsa.PrivateKey
	Client                     *http.Client

	mu     sync.Mutex
	token  string
	signed time.Time
}

// ParseAppleKey reads the push key as the .p8 file in base64, one line as a .env value needs; "" is no key.
func ParseAppleKey(b64 string) (*ecdsa.PrivateKey, error) {
	if b64 == "" {
		return nil, nil
	}
	p8, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("the Apple push key is not base64: %w", err)
	}
	block, _ := pem.Decode(p8)
	if block == nil {
		return nil, errors.New("the Apple push key is not a .p8 file")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("the Apple push key is not an EC key")
	}
	return ec, nil
}

// errGone is Apple's 410: the device token is no longer valid for the topic.
var errGone = errors.New("apple: the device token is gone")

// send is one request for one device over HTTP/2, carrying body as the alert.
func (a *Apple) send(ctx context.Context, device, body string) error {
	token, err := a.bearer()
	if err != nil {
		return err
	}
	payload, err := json.Marshal(map[string]any{"aps": map[string]any{"alert": map[string]string{"body": body}}})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.Host+"/3/device/"+device, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("authorization", "bearer "+token)
	req.Header.Set("apns-topic", a.Topic)
	req.Header.Set("apns-push-type", "alert")
	res, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	reason, _ := io.ReadAll(io.LimitReader(res.Body, 512))
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusGone:
		return errGone
	}
	return fmt.Errorf("apple answered %d %s", res.StatusCode, reason)
}

// bearer is the ES256 token of the team and Key ID, signed again once tokenLife has passed.
func (a *Apple) bearer() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.token != "" && time.Since(a.signed) < tokenLife {
		return a.token, nil
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: a.Key}, (&jose.SignerOptions{}).WithHeader("kid", a.KeyID))
	if err != nil {
		return "", err
	}
	now := time.Now()
	token, err := jwt.Signed(signer).Claims(jwt.Claims{Issuer: a.TeamID, IssuedAt: jwt.NewNumericDate(now)}).Serialize()
	if err != nil {
		return "", err
	}
	a.token, a.signed = token, now
	return token, nil
}

// warningKind is the Target Warning's kind in sent_log, named as its switch's column; the period is the month.
const warningKind = "target_warning"

// pushWarning pushes a month's first stamp, once committed, in the background (ARCHITECTURE §4.3).
func (s *Service) pushWarning(ctx context.Context, id, month string) {
	s.pushes.Go(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), pushTimeout)
		defer cancel()
		logFailed(warningKind, month, s.warn(ctx, id, month))
	})
}

func (s *Service) warn(ctx context.Context, id, month string) error {
	at, err := time.Parse("2006-01", month)
	if err != nil {
		return err
	}
	return s.push(ctx, id, warningKind, month, "You've passed your saving line for "+at.Format("January")+".")
}

// push sends body to every device of the Member unless kind's switch, its column in notification_preference, is off
// or a deletion is pending; sent_log is claimed first, so a crash loses the push and a second run never resends it.
func (s *Service) push(ctx context.Context, id, kind, period, body string) error {
	claimed, err := s.DB.Exec(ctx, `INSERT INTO sent_log (member_id, kind, period)
		SELECT $1, $2, $3
		WHERE coalesce((SELECT `+kind+` FROM notification_preference WHERE member_id = $1), $4)
			AND NOT coalesce((SELECT deletion_pending FROM member_facts WHERE member_id = $1), false)
		ON CONFLICT DO NOTHING`, id, kind, period, kind != reminderKind)
	if err != nil || claimed.RowsAffected() == 0 {
		return err
	}
	if s.Apple == nil || s.Apple.Key == nil {
		log.Printf("notify: no Apple push key, so the %s push for %s is recorded and not sent", kind, period)
		return nil
	}
	rows, err := s.DB.Query(ctx, `SELECT apns_token FROM device WHERE member_id = $1`, id)
	if err != nil {
		return err
	}
	devices, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	var failed error
	for _, device := range devices {
		err := s.Apple.send(ctx, device, body)
		if errors.Is(err, errGone) {
			_, err = s.DB.Exec(ctx, `DELETE FROM device WHERE apns_token = $1 AND member_id = $2`, device, id)
		}
		failed = errors.Join(failed, err)
	}
	return failed
}
