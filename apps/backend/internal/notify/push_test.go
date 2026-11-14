package notify

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
)

// push is one request the fake Apple service received.
type push struct {
	path, topic, pushType, token, body string
	proto                              int
}

// fakeApple is Apple's push service over TLS and HTTP/2, answering 410 for the device tokens in gone.
type fakeApple struct {
	mu     sync.Mutex
	pushes []push
	gone   map[string]bool
}

func (f *fakeApple) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pushes = append(f.pushes, push{r.URL.Path, r.Header.Get("apns-topic"), r.Header.Get("apns-push-type"), strings.TrimPrefix(r.Header.Get("authorization"), "bearer "), string(body), r.ProtoMajor})
	if f.gone[strings.TrimPrefix(r.URL.Path, "/3/device/")] {
		w.WriteHeader(http.StatusGone)
		_, _ = w.Write([]byte(`{"reason":"Unregistered"}`))
	}
}

func (f *fakeApple) received() []push {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]push(nil), f.pushes...)
}

// pushService is warningService sending to a fake Apple service with a key generated here, the Member's devices,
// preferences and sent_log removed at the end.
func pushService(t *testing.T, id string, gone ...string) (*Service, *fakeApple, *ecdsa.PrivateKey) {
	t.Helper()
	s := warningService(t, id, map[string]string{"owner": id})
	apple, fake, key := startApple(t, gone...)
	s.Apple = apple
	t.Cleanup(func() {
		s.pushes.Wait()
		for _, table := range []string{"device", "notification_preference"} {
			if _, err := s.DB.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = $1`, id); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
	return s, fake, key
}

// startApple is a fake Apple push service over TLS and HTTP/2, answering 410 for gone, and a sender to it signing
// with a key generated here.
func startApple(t *testing.T, gone ...string) (*Apple, *fakeApple, *ecdsa.PrivateKey) {
	t.Helper()
	fake := &fakeApple{gone: map[string]bool{}}
	for _, token := range gone {
		fake.gone[token] = true
	}
	srv := httptest.NewUnstartedServer(fake)
	srv.EnableHTTP2 = true
	srv.StartTLS()
	t.Cleanup(srv.Close)
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return &Apple{Host: srv.URL, Topic: "com.mymoneyrecap.app", TeamID: "TEAM123456", KeyID: "KEY1234567", Key: key, Client: srv.Client()}, fake, key
}

func devices(t *testing.T, s *Service, id string, tokens ...string) {
	t.Helper()
	for _, token := range tokens {
		if _, err := s.DB.Exec(t.Context(), `INSERT INTO device (apns_token, member_id) VALUES ($1, $2)`, token, id); err != nil {
			t.Fatal(err)
		}
	}
}

// cross calls EvaluateWarning over the line for month and waits for any push it started.
func cross(t *testing.T, s *Service, id, key, month string) {
	t.Helper()
	_, err := s.EvaluateWarning(as("owner"), connect.NewRequest(&moneyrecapv1.EvaluateWarningRequest{
		IdempotencyKey: key, MemberId: id, Month: month, CrossedLine: true, Warning: &moneyrecapv1.TargetWarning{Message: "over"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	s.pushes.Wait()
}

// crossByEvent hands the consumer one crossing Spend event for month, as the broker would, and waits for any push.
func crossByEvent(t *testing.T, s *Service, id, messageID, month string) {
	t.Helper()
	body, err := proto.Marshal(&moneyrecapv1.LedgerSpendLogged{
		Envelope: &moneyrecapv1.EventEnvelope{MemberId: id}, CrossedLine: true,
		Month: &moneyrecapv1.MonthSummary{Month: month, TargetWarning: &moneyrecapv1.TargetWarning{Message: "over"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := s.DB.Exec(context.Background(), `DELETE FROM processed_events WHERE message_id = $1`, messageID); err != nil {
			t.Logf("cleanup processed_events: %v", err)
		}
	})
	a := &acks{}
	events.Process(t.Context(), s.DB, amqp.Delivery{Acknowledger: a, MessageId: messageID, RoutingKey: Keys[0], Body: body}, s.StampFromEvent)
	if a.acked != 1 {
		t.Fatalf("the event for %s was not acked", month)
	}
	s.pushes.Wait()
}

// A month's first stamp pushes PRD §10's copy to each device over HTTP/2, signed for the team and Key ID; a 410
// deletes the device, and no later stamp by either path or replica pushes again, while the next month's does.
func TestTheFirstStampOfAMonthPushesOnce(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000f001"
	phone, pad := "aa01", "aa02"
	s, fake, key := pushService(t, id, pad)
	devices(t, s, id, phone, pad)

	cross(t, s, id, "k1", "2026-09")
	got := fake.received()
	if len(got) != 2 {
		t.Fatalf("Apple received %d pushes, want one per device", len(got))
	}
	for _, p := range got {
		if p.path != "/3/device/"+phone && p.path != "/3/device/"+pad {
			t.Errorf("a push to %s, want one of the Member's devices", p.path)
		}
		if p.proto != 2 || p.topic != "com.mymoneyrecap.app" || p.pushType != "alert" {
			t.Errorf("a push over HTTP/%d with topic %q and push type %q, want HTTP/2, the app's topic and alert", p.proto, p.topic, p.pushType)
		}
		var payload struct {
			Aps struct {
				Alert struct{ Body string } `json:"alert"`
			} `json:"aps"`
		}
		if err := json.Unmarshal([]byte(p.body), &payload); err != nil || payload.Aps.Alert.Body != "You've passed your saving line for September." {
			t.Errorf("the push carried %s (%v), want the September copy as its alert body", p.body, err)
		}
		signed, err := jwt.ParseSigned(p.token, []jose.SignatureAlgorithm{jose.ES256})
		if err != nil {
			t.Fatalf("the bearer token: %v", err)
		}
		var claims jwt.Claims
		if err := signed.Claims(&key.PublicKey, &claims); err != nil || claims.Issuer != "TEAM123456" || signed.Headers[0].KeyID != "KEY1234567" {
			t.Errorf("the token: %v, issuer %q, kid %q, want signed by the push key for the team and its Key ID", err, claims.Issuer, signed.Headers[0].KeyID)
		}
		if claims.IssuedAt == nil || time.Since(claims.IssuedAt.Time()) > time.Minute {
			t.Errorf("the token was issued at %v, want now", claims.IssuedAt)
		}
	}
	var held []string
	rows, _ := s.DB.Query(t.Context(), `SELECT apns_token FROM device WHERE member_id = $1`, id)
	for rows.Next() {
		var token string
		_ = rows.Scan(&token)
		held = append(held, token)
	}
	if rows.Err() != nil || len(held) != 1 || held[0] != phone {
		t.Errorf("after Apple's 410 the Member holds %v (%v), want the phone alone", held, rows.Err())
	}

	cross(t, s, id, "k2", "2026-09")
	crossByEvent(t, s, id, "push-f001-a", "2026-09")
	other := &Service{DB: s.DB, Apple: s.Apple}
	other.pushWarning(t.Context(), id, "2026-09")
	other.pushes.Wait()
	if n := len(fake.received()); n != 2 {
		t.Fatalf("after later stamps of September, a second replica included, Apple received %d pushes, want still 2", n)
	}
	crossByEvent(t, s, id, "push-f001-b", "2026-10")
	crossByEvent(t, s, id, "push-f001-c", "2026-10")
	cross(t, s, id, "k3", "2026-10")
	got = fake.received()[2:]
	if len(got) != 1 || !strings.Contains(got[0].body, "saving line for October.") {
		t.Errorf("October's stamps pushed %d times (%v), want once, to the phone, naming October", len(got), got)
	}
}

// No push goes out while the Target Warning switch is off or a deletion is pending, and the month is stamped all the
// same, so Home's banner is unaffected (spec #169, "What every push obeys").
func TestNoPushWhenTheSwitchIsOffOrADeletionIsPending(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000f002"
	s, fake, _ := pushService(t, id)
	devices(t, s, id, "bb01")
	if _, err := s.DB.Exec(t.Context(), `INSERT INTO notification_preference VALUES ($1, true, false, '20:00', false)`, id); err != nil {
		t.Fatal(err)
	}
	cross(t, s, id, "k1", "2026-09")
	if _, err := s.DB.Exec(t.Context(), `DELETE FROM notification_preference WHERE member_id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if _, err := s.DB.Exec(t.Context(), `UPDATE member_facts SET deletion_pending = true WHERE member_id = $1`, id); err != nil {
		t.Fatal(err)
	}
	crossByEvent(t, s, id, "push-f002-a", "2026-10")
	if n := len(fake.received()); n != 0 {
		t.Errorf("Apple received %d pushes, want none with the switch off or a deletion pending", n)
	}
	for _, month := range []string{"2026-09", "2026-10"} {
		res, err := s.GetWarningState(as("owner"), connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: id, Month: month}))
		if err != nil || !res.Msg.Fired {
			t.Errorf("%s: %v, %v, want stamped for the banner", month, res, err)
		}
	}
}

// With no key configured, as on every laptop, the push is recorded in sent_log and skipped with one log line.
func TestNoKeyRecordsThePushAndSkipsIt(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000f003"
	s, fake, _ := pushService(t, id)
	s.Apple.Key = nil
	devices(t, s, id, "cc01")
	var logged bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(prev) })
	cross(t, s, id, "k1", "2026-09")
	log.SetOutput(prev)
	var recorded int
	if err := s.DB.QueryRow(t.Context(), `SELECT count(*) FROM sent_log WHERE member_id = $1 AND kind = 'target_warning' AND period = '2026-09'`, id).Scan(&recorded); err != nil || recorded != 1 {
		t.Errorf("sent_log holds %d rows (%v), want the push recorded", recorded, err)
	}
	if n := len(fake.received()); n != 0 {
		t.Errorf("Apple received %d pushes with no key, want none", n)
	}
	if lines := strings.Count(logged.String(), "\n"); lines != 1 || !strings.Contains(logged.String(), "no Apple push key") {
		t.Errorf("logged %q, want one line saying no key is configured", logged.String())
	}
}

// The push key arrives as the .p8 file in base64; an empty value is no key, and anything else that is not an EC
// .p8 stops the start-up.
func TestParseAppleKey(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	p8 := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	if got, err := ParseAppleKey(p8); err != nil || !got.Equal(key) {
		t.Errorf("a .p8 in base64: %v, want the key", err)
	}
	if got, err := ParseAppleKey(""); err != nil || got != nil {
		t.Errorf("no value: %v, %v, want no key and no error", got, err)
	}
	for name, value := range map[string]string{
		"not base64": "%%%",
		"not a PEM":  base64.StdEncoding.EncodeToString([]byte("replace-me")),
		"not PKCS8":  base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte{1}})),
	} {
		if _, err := ParseAppleKey(value); err == nil {
			t.Errorf("%s: parsed, want an error", name)
		}
	}
}
