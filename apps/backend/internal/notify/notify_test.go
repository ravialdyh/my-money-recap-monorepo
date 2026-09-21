package notify

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
)

// A missing key would dedupe every later send into silence, so it is refused before the store is
// touched (#68); the zero Service proves it never reaches a pool.
func TestAnEmptyIdempotencyKeyIsRefused(t *testing.T) {
	_, err := (&Service{}).EmailMagicLink(context.Background(), connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
		Email: "someone@example.test", Link: "https://mymoneyrecap.test/magic?lt=x",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("empty idempotency key: %v, want invalid_argument", err)
	}
}

// testDB opens the Compose stack's notify database, skipping unless the NOTIFY_TEST_* variables
// infra/compose/README.md lists are set.
func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTIFY_TEST_DATABASE_URL not set")
	}
	db, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return db
}

// TestMagicLinkAtTheConnectBoundary drives the running service the way the BFF will: no bearer token
// at all, which only the EmailMagicLink exemption lets through (ARCHITECTURE §4.2 point 3).
func TestMagicLinkAtTheConnectBoundary(t *testing.T) {
	base, mailpit := os.Getenv("NOTIFY_TEST_URL"), os.Getenv("NOTIFY_TEST_MAILPIT")
	if base == "" || mailpit == "" {
		t.Skip("NOTIFY_TEST_URL and NOTIFY_TEST_MAILPIT not both set")
	}
	db := testDB(t)
	ctx := t.Context()
	address := "mmr-test-" + strings.ToLower(rand.Text()) + "@example.test"
	first, second := strings.ToLower(rand.Text()), strings.ToLower(rand.Text())
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM magic_link_send WHERE email = $1`, address); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})

	client := moneyrecapv1connect.NewNotifyServiceClient(http.DefaultClient, base)
	send := func(key, link string) error {
		_, err := client.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
			IdempotencyKey: key, Email: address, Link: link,
		}))
		return err
	}
	link := "https://mymoneyrecap.test/magic?lt=" + first
	if err := send(first, link); err != nil {
		t.Fatalf("first call without a bearer token: %v", err)
	}
	if err := send(first, link+"-retry"); err != nil {
		t.Fatalf("second call with the same key: %v, want success", err)
	}

	inbox := waitForMail(t, mailpit, address)
	if len(inbox) != 1 {
		t.Fatalf("%d messages for one idempotency key, want 1", len(inbox))
	}
	if body := messageText(t, mailpit, inbox[0]); !strings.Contains(body, link) {
		t.Fatalf("message body %q carries no %q", body, link)
	}
	t.Logf("two calls, one key → 1 message in Mailpit, carrying %s", link)

	var stored string
	var sentAt time.Time
	if err := db.QueryRow(ctx, `SELECT email, sent_at FROM magic_link_send WHERE idempotency_key = $1`, first).Scan(&stored, &sentAt); err != nil {
		t.Fatal(err)
	}
	if stored != address || time.Since(sentAt) > time.Minute {
		t.Fatalf("row: %q at %s, want %q just now", stored, sentAt, address)
	}
	t.Logf("magic_link_send row: key %s, %s, %s", first, stored, sentAt.Format(time.RFC3339))

	if err := send(second, link+"-again"); err != nil {
		t.Fatalf("a second key to the same address: %v", err)
	}
	if inbox := waitForMail(t, mailpit, address); len(inbox) != 2 {
		t.Fatalf("%d messages after a second key, want 2: the key dedupes, not the address", len(inbox))
	}
	t.Logf("a second key to %s → 2 messages", address)

	if err := send("", link); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("empty key over the wire: %v, want invalid_argument", err)
	}
	_, err := client.GetPrefs(ctx, connect.NewRequest(&moneyrecapv1.GetPrefsRequest{MemberId: "00000000-0000-0000-0000-000000000001"}))
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("GetPrefs without a token: %v, want unauthenticated: the exemption is one procedure, not the service", err)
	}
	t.Logf("GetPrefs without a token: %v", err)
}

// waitForMail is every message Mailpit holds for address, polled for three seconds so the first one
// is never missed to storage lag; a later duplicate is already stored, because the send happens
// before EmailMagicLink answers.
func waitForMail(t *testing.T, mailpit, address string) []string {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		var found struct {
			Messages []struct{ ID string }
		}
		get(t, mailpit+"/api/v1/search?query="+url.QueryEscape("to:"+address), &found)
		ids := make([]string, 0, len(found.Messages))
		for _, m := range found.Messages {
			ids = append(ids, m.ID)
		}
		if len(ids) > 0 || time.Now().After(deadline) {
			return ids
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func messageText(t *testing.T, mailpit, id string) string {
	t.Helper()
	var message struct{ Text string }
	get(t, mailpit+"/api/v1/message/"+id, &message)
	return message.Text
}

// get decodes Mailpit's answer into out, failing the test on anything but a 200 of JSON.
func get(t *testing.T, address string, out any) {
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

// The row is claimed before the send and a refused SMTP host rolls it back, so a retry of the same
// key answers success without a second email and a failed send never spends the key (A14).
func TestAClaimedKeySendsOnceAndAFailedSendFreesIt(t *testing.T) {
	host := os.Getenv("NOTIFY_TEST_SMTP")
	if host == "" {
		t.Skip("NOTIFY_TEST_SMTP not set")
	}
	db := testDB(t)
	ctx := t.Context()
	if err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	address := "mmr-test-" + strings.ToLower(rand.Text()) + "@example.test"
	sent, refused := strings.ToLower(rand.Text()), strings.ToLower(rand.Text())
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM magic_link_send WHERE email = $1`, address); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	send := func(s *Service, key string) error {
		_, err := s.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
			IdempotencyKey: key, Email: address, Link: "https://mymoneyrecap.test/magic?lt=" + key,
		}))
		return err
	}
	rows := func(key string) int {
		var n int
		if err := db.QueryRow(ctx, `SELECT count(*) FROM magic_link_send WHERE idempotency_key = $1`, key).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	mailpit := &Service{DB: db, SMTP: host, From: "no-reply@mymoneyrecap.test"}
	if err := send(mailpit, sent); err != nil {
		t.Fatal(err)
	}
	if err := send(mailpit, sent); err != nil || rows(sent) != 1 {
		t.Fatalf("the same key twice: %v, %d rows, want success and 1", err, rows(sent))
	}

	nowhere := &Service{DB: db, SMTP: "127.0.0.1:1", From: "no-reply@mymoneyrecap.test"}
	err := send(nowhere, refused)
	if connect.CodeOf(err) != connect.CodeUnavailable || strings.Contains(err.Error(), "127.0.0.1") {
		t.Fatalf("a refused SMTP host: %v, want unavailable without its address", err)
	}
	if rows(refused) != 0 {
		t.Fatalf("%d rows after a refused send, want the key free", rows(refused))
	}
	if err := send(mailpit, refused); err != nil || rows(refused) != 1 {
		t.Fatalf("the freed key sent again: %v, %d rows, want success and 1", err, rows(refused))
	}
	t.Logf("key %s claimed once; key %s rolled back by a refused host, then spent", sent, refused)
}
