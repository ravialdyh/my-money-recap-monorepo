package notify

import (
	"bufio"
	"context"
	"crypto/rand"
	"io"
	"net"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

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

func TestSendsOlderThanTheRetentionWindowAreSwept(t *testing.T) {
	host := os.Getenv("NOTIFY_TEST_SMTP")
	if host == "" {
		t.Skip("NOTIFY_TEST_SMTP not set")
	}
	db := testDB(t)
	ctx := t.Context()
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
