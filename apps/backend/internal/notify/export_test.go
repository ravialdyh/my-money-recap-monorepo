package notify

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// recordingSMTP is a mail host on a free port that accepts every message and keeps each one's DATA.
func recordingSMTP(t *testing.T) (addr string, sent func() []string) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	var mu sync.Mutex
	var kept []string
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				r := bufio.NewReader(conn)
				_, _ = conn.Write([]byte("220 fake ESMTP\r\n"))
				for {
					line, err := r.ReadString('\n')
					if err != nil {
						return
					}
					switch strings.ToUpper(strings.TrimSpace(line)) {
					case "DATA":
						_, _ = conn.Write([]byte("354 go\r\n"))
						var data strings.Builder
						for {
							l, err := r.ReadString('\n')
							if err != nil || l == ".\r\n" {
								break
							}
							data.WriteString(l)
						}
						mu.Lock()
						kept = append(kept, data.String())
						mu.Unlock()
						_, _ = conn.Write([]byte("250 queued\r\n"))
					case "QUIT":
						_, _ = conn.Write([]byte("221 bye\r\n"))
						return
					default:
						_, _ = conn.Write([]byte("250 ok\r\n"))
					}
				}
			}()
		}
	}()
	return listener.Addr().String(), func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), kept...)
	}
}

// The export goes to the replica's address with spec #169's subject, body and attachment, once per key: a retry sends
// nothing; a refused address is failed_precondition naming it, any other failure unavailable; neither spends the key.
func TestTheExportIsMailedOncePerKey(t *testing.T) {
	ann := "00000000-0000-4000-8000-00000000e182"
	s := warningService(t, ann, map[string]string{"ann": ann})
	t.Cleanup(func() {
		if _, err := s.DB.Exec(context.Background(), `DELETE FROM idempotency WHERE member_id = $1`, ann); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	s.From = "no-reply@mymoneyrecap.test"
	csv := "date,category,amount,currency,note\r\n2026-07-10,Snacks,4.50,GBP,Kopi\r\n"
	export := func(key string) error {
		_, err := s.EmailExport(as("ann"), connect.NewRequest(&moneyrecapv1.EmailExportRequest{IdempotencyKey: key, MemberId: ann, Csv: []byte(csv)}))
		return err
	}

	s.SMTP = fakeSMTP(t, "220 fake ESMTP", "250 fake", "250 2.1.0 Ok", "550 5.1.1 Recipient address rejected")
	err := export("k1")
	var bounced *moneyrecapv1.EmailBounced
	if ce := new(connect.Error); connect.CodeOf(err) == connect.CodeFailedPrecondition && errors.As(err, &ce) {
		for _, d := range ce.Details() {
			if v, derr := d.Value(); derr == nil {
				bounced, _ = v.(*moneyrecapv1.EmailBounced)
			}
		}
	}
	if bounced.GetAddress() != "ann@example.test" || bounced.GetBouncedAt() == nil {
		t.Fatalf("a refused address: %v, want failed_precondition with EmailBounced naming it", err)
	}
	s.SMTP = fakeSMTP(t, "220 fake ESMTP", "250 fake", "250 Ok", "250 Ok", "354 go", "554 5.3.0 Message rejected")
	if err := export("k1"); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("a rejected message: %v, want unavailable", err)
	}

	host, sent := recordingSMTP(t)
	s.SMTP = host
	for range 2 {
		if err := export("k1"); err != nil {
			t.Fatalf("k1 after two failures: %v, want sent, the key still free", err)
		}
	}
	if n := len(sent()); n != 1 {
		t.Fatalf("%d messages for one key sent twice, want 1", n)
	}
	msg, err := mail.ReadMessage(strings.NewReader(sent()[0]))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Header.Get("To") != "ann@example.test" || msg.Header.Get("Subject") != "Your spends from My Money Recap" {
		t.Errorf("headers %v, want the replica's address and the spec's subject", msg.Header)
	}
	_, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	parts := multipart.NewReader(msg.Body, params["boundary"])
	text, err := parts.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if body, _ := io.ReadAll(text); strings.TrimSpace(string(body)) != "Here is every spend you've logged, as a CSV file." {
		t.Errorf("body %q, want the spec's line", body)
	}
	file, err := parts.NextRawPart()
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := io.ReadAll(file)
	attached, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(string(encoded), "\r\n", ""))
	if err != nil || file.FileName() != "my-money-recap-spends.csv" || string(attached) != csv {
		t.Errorf("attachment %q, want my-money-recap-spends.csv holding the CSV as given", file.FileName())
	}
	if err := export(""); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no key: %v, want invalid_argument", err)
	}
}
