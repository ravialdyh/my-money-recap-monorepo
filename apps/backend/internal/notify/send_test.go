package notify

import (
	"bufio"
	"context"
	"crypto/rand"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// fakeSMTP is a mail host on a free port that answers each command with the next reply and then
// reads without answering, so no reply at all is a host that accepts the connection and says nothing.
func fakeSMTP(t *testing.T, replies ...string) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go serveFakeSMTP(conn, replies)
		}
	}()
	return listener.Addr().String()
}

// serveFakeSMTP writes a reply, reads the command it drew, and drains the rest in silence.
func serveFakeSMTP(conn net.Conn, replies []string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for _, reply := range replies {
		if _, err := conn.Write([]byte(reply + "\r\n")); err != nil {
			return
		}
		if _, err := reader.ReadString('\n'); err != nil {
			return
		}
	}
	_, _ = io.Copy(io.Discard, reader)
}

// testPool is the Compose stack's notify database on a pool of maxConns, so a test can run the pool
// out of connections the way two wedged sends do.
func testPool(t *testing.T, maxConns int32) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTIFY_TEST_DATABASE_URL not set")
	}
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.MaxConns = maxConns
	db, err := pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return db
}

// A mail host that accepts and then goes silent must not outlive the caller's deadline: it holds a
// pool connection idle in transaction, which starves the db.Ping behind /healthz and blocks the
// db.Close cmd/notify defers.
func TestASilentMailHostCannotOutliveTheCallersDeadline(t *testing.T) {
	db := testPool(t, 2)
	service := &Service{DB: db, SMTP: fakeSMTP(t), From: "no-reply@mymoneyrecap.test"}
	codes := make(chan connect.Code, 2)
	var sends sync.WaitGroup
	for range cap(codes) {
		sends.Add(1)
		go func() {
			defer sends.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err := service.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
				IdempotencyKey: strings.ToLower(rand.Text()),
				Email:          "mmr-test-" + strings.ToLower(rand.Text()) + "@example.test",
				Link:           "https://mymoneyrecap.test/magic?lt=x",
			}))
			codes <- connect.CodeOf(err)
		}()
	}
	waitForAcquired(t, db, int32(cap(codes)))

	ping, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := db.Ping(ping); err != nil {
		t.Fatalf("/healthz while two sends are in flight: %v, want the pool free within 5s", err)
	}
	sends.Wait()
	for range cap(codes) {
		if code := <-codes; code != connect.CodeUnavailable {
			t.Errorf("a silent mail host: %v, want unavailable", code)
		}
	}
	closed := time.Now()
	db.Close()
	if elapsed := time.Since(closed); elapsed > time.Second {
		t.Fatalf("db.Close took %s, want prompt: main defers it and the container must shut down", elapsed)
	}
	t.Logf("two sends released the pool inside their 2s deadline; db.Close returned in %s", time.Since(closed))
}

// waitForAcquired blocks until the pool has handed out want connections, so the /healthz assertion
// runs while the sends really hold them rather than racing their first statement.
func waitForAcquired(t *testing.T, db *pgxpool.Pool, want int32) {
	t.Helper()
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if db.Stat().AcquiredConns() >= want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%d connections acquired, want %d", db.Stat().AcquiredConns(), want)
}

// A permanent 5xx on the recipient is a wrong address, not a mail host to retry: it answers
// invalid_argument, so a mistyped address is told apart and a retry does not spend one of the
// five-per-address-per-hour slots #69 adds. The key stays free.
func TestAPermanentRecipientRefusalIsInvalidArgument(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()
	key := strings.ToLower(rand.Text())
	refusing := fakeSMTP(t, "220 fake ESMTP", "250 fake", "250 2.1.0 Ok", "550 5.1.1 Recipient address rejected")
	service := &Service{DB: db, SMTP: refusing, From: "no-reply@mymoneyrecap.test"}
	_, err := service.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
		IdempotencyKey: key, Email: "nobody@example.test", Link: "https://mymoneyrecap.test/magic?lt=x",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a permanent 550 on RCPT TO: %v, want invalid_argument", err)
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM magic_link_send WHERE idempotency_key = $1`, key).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("%d rows after a refused recipient, want the key free", rows)
	}
	t.Logf("550 5.1.1 → %v, key %s unspent", connect.CodeOf(err), key)
}

// Only the recipient's refusal is the caller's fault. A host that cannot be dialled, one that rejects
// the envelope sender and one that rejects the message body are all the mail host failing, so they
// stay unavailable and the caller may retry.
func TestEveryOtherSMTPFailureStaysUnavailable(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()
	for name, host := range map[string]string{
		"no port in the mail host": "mailpit-without-a-port",
		"the sender is rejected":   fakeSMTP(t, "220 fake ESMTP", "250 fake", "550 5.7.1 Sender rejected"),
		"the body is rejected":     fakeSMTP(t, "220 fake ESMTP", "250 fake", "250 Ok", "250 Ok", "354 go", "554 5.3.0 Message rejected"),
	} {
		t.Run(name, func(t *testing.T) {
			service := &Service{DB: db, SMTP: host, From: "no-reply@mymoneyrecap.test"}
			_, err := service.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
				IdempotencyKey: strings.ToLower(rand.Text()),
				Email:          "mmr-test-" + strings.ToLower(rand.Text()) + "@example.test",
				Link:           "https://mymoneyrecap.test/magic?lt=x",
			}))
			if connect.CodeOf(err) != connect.CodeUnavailable {
				t.Fatalf("%s: %v, want unavailable", name, err)
			}
			if strings.Contains(err.Error(), host) {
				t.Fatalf("%v names the mail host, which the caller may not see", err)
			}
		})
	}
}

// An address carrying CR or LF is refused before any recipient is accepted, which is the whole of the
// SMTP header-injection defence on this unauthenticated procedure (#61's pre-authentication path).
func TestAnAddressCarryingCRLFIsRefused(t *testing.T) {
	db := testDB(t)
	ctx := t.Context()
	key := strings.ToLower(rand.Text())
	host := fakeSMTP(t, "220 fake ESMTP", "250 fake", "250 2.1.0 Ok", "250 2.1.5 Ok", "354 go", "250 2.0.0 Ok")
	service := &Service{DB: db, SMTP: host, From: "no-reply@mymoneyrecap.test"}
	_, err := service.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
		IdempotencyKey: key,
		Email:          "victim@example.test\r\nBcc: attacker@evil.test",
		Link:           "https://mymoneyrecap.test/magic?lt=x",
	}))
	if err == nil {
		t.Fatal("an address holding CRLF was accepted, want it refused")
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM magic_link_send WHERE idempotency_key = $1`, key).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("%d rows after a refused address, want none", rows)
	}
	t.Logf("CRLF in the address → %v, nothing stored", connect.CodeOf(err))
}

// A send older than the retention window goes on the next send and a fresh one stays: the row holds a
// raw address and no member_id, so ARCHITECTURE §7.3's members.deleted purge can never reach it.
func TestSendsOlderThanTheRetentionWindowAreSwept(t *testing.T) {
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
	stale, fresh, now := strings.ToLower(rand.Text()), strings.ToLower(rand.Text()), strings.ToLower(rand.Text())
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM magic_link_send WHERE email = $1`, address); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	seed := `INSERT INTO magic_link_send (idempotency_key, email, sent_at) VALUES ($1, $2, $3)`
	if _, err := db.Exec(ctx, seed, stale, address, time.Now().Add(-retention-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, seed, fresh, address, time.Now()); err != nil {
		t.Fatal(err)
	}
	service := &Service{DB: db, SMTP: host, From: "no-reply@mymoneyrecap.test"}
	if _, err := service.EmailMagicLink(ctx, connect.NewRequest(&moneyrecapv1.EmailMagicLinkRequest{
		IdempotencyKey: now, Email: address, Link: "https://mymoneyrecap.test/magic?lt=" + now,
	})); err != nil {
		t.Fatal(err)
	}
	var kept []string
	if err := db.QueryRow(ctx, `SELECT array_agg(idempotency_key ORDER BY idempotency_key) FROM magic_link_send WHERE email = $1`, address).Scan(&kept); err != nil {
		t.Fatal(err)
	}
	want := slices.Sorted(slices.Values([]string{fresh, now}))
	if !slices.Equal(kept, want) {
		t.Fatalf("rows %v, want %v: the %s-old key %s should be swept", kept, want, retention+time.Hour, stale)
	}
	t.Logf("a row %s old was swept by the next send; the fresh one stayed", retention+time.Hour)
}

// waitForMail must wait for the count its caller asks for: it returned on the first message it saw,
// so the two-message assertion in TestMagicLinkAtTheConnectBoundary did no waiting at all.
func TestWaitForMailWaitsForTheCountItIsGiven(t *testing.T) {
	var polls atomic.Int32
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		found := `{"messages":[{"ID":"one"}]}`
		if polls.Add(1) > 2 {
			found = `{"messages":[{"ID":"one"},{"ID":"two"}]}`
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(found))
	}))
	t.Cleanup(stub.Close)
	if inbox := waitForMail(t, stub.URL, "someone@example.test", 2); len(inbox) != 2 {
		t.Fatalf("%d messages, want 2: the helper answered before the stub held them both", len(inbox))
	}
	t.Logf("waitForMail polled %d times for 2 messages", polls.Load())
}
