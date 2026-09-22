// Package notify is the notify service (ARCHITECTURE §7.3). Since #68 it sends the magic-link
// email and nothing else: its schema, its SMTP client and its Connect handler.
package notify

import (
	"context"
	"crypto/tls"
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"net/textproto"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/migrate"
)

//go:embed migrations/*.sql
var migrations embed.FS

// retention bounds how long a send row lives. Its link dies in 15 minutes (PRD D32), so a day-old
// key has nothing left to dedupe, and the raw address it holds has no member_id for §7.3's purge to
// reach. Every send sweeps, so no scheduler is owed.
const retention = 24 * time.Hour

// sendTimeout caps one SMTP conversation when the caller set no shorter deadline: a mail host that
// accepts the connection and then says nothing must not hold a pool connection, and the transaction
// open on it, for ever.
const sendTimeout = 10 * time.Second

// errAddressRefused is a permanent 5xx on the recipient: the address is wrong, not the mail host, so
// the caller hears invalid_argument and a retry does not spend one of its five-per-hour slots.
var errAddressRefused = errors.New("the mail host refused the address")

// Migrate applies this service's embedded migrations to the notify database (internal/migrate).
func Migrate(ctx context.Context, db *pgxpool.Pool) error { return migrate.Run(ctx, db, migrations) }

// Service implements NotifyService. It serves EmailMagicLink alone (#68), which is the one procedure
// outside the bearer check (internal/auth); every other RPC arrives in slice 7 behind it.
type Service struct {
	moneyrecapv1connect.UnimplementedNotifyServiceHandler
	DB *pgxpool.Pool
	// SMTP is the mail host as host:port: Mailpit on a laptop and in CI, a provider on the VM (§7.5).
	SMTP string
	// From is the envelope sender and the From header.
	From string
}

// EmailMagicLink sends the link once per address and idempotency key: the row is claimed first and a
// failed send rolls it back, so a repeated key answers success without a second email (A14). An empty
// key would dedupe every later send into silence.
func (s *Service) EmailMagicLink(ctx context.Context, req *connect.Request[moneyrecapv1.EmailMagicLinkRequest]) (*connect.Response[moneyrecapv1.EmailMagicLinkResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM magic_link_send WHERE sent_at < $1`, time.Now().Add(-retention)); err != nil {
			return err
		}
		claimed, err := tx.Exec(ctx, `INSERT INTO magic_link_send (idempotency_key, email) VALUES ($1, $2) ON CONFLICT (email, idempotency_key) DO NOTHING`, req.Msg.IdempotencyKey, req.Msg.Email)
		if err != nil || claimed.RowsAffected() == 0 {
			return err
		}
		return s.send(ctx, req.Msg.Email, req.Msg.Link)
	})
	if err != nil {
		log.Printf("notify magic link: %v", err)
		if errors.Is(err, errAddressRefused) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errAddressRefused)
		}
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("the magic-link email was not sent"))
	}
	return connect.NewResponse(&moneyrecapv1.EmailMagicLinkResponse{}), nil
}

// send opens the connection under ctx or sendTimeout, whichever is sooner, and puts that deadline on
// every read and write: a connection accepted and then left silent would outlive the caller, holding
// its transaction and blocking the pool's Close.
func (s *Service) send(ctx context.Context, to, link string) error {
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	host, _, err := net.SplitHostPort(s.SMTP)
	if err != nil {
		return err
	}
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", s.SMTP)
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	return s.deliver(client, host, to, link)
}

// deliver is the one exchange, its body the single-use link of PRD D32. Client.Mail and Client.Rcpt
// refuse an address holding CR or LF, the whole header-injection defence on this unauthenticated
// path; STARTTLS is taken whenever the host offers it.
func (s *Service) deliver(client *smtp.Client, host, to, link string) error {
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if err := client.Mail(s.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		var refusal *textproto.Error
		if errors.As(err, &refusal) && refusal.Code >= 500 {
			return fmt.Errorf("%w: %w", errAddressRefused, err)
		}
		return err
	}
	body, err := client.Data()
	if err != nil {
		return err
	}
	msg := "From: " + s.From + "\r\nTo: " + to +
		"\r\nSubject: Your My Money Recap sign-in link\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n\r\n" +
		"Tap this link to sign in. It works for 15 minutes, once.\r\n\r\n" + link + "\r\n"
	if _, err := body.Write([]byte(msg)); err != nil {
		return err
	}
	if err := body.Close(); err != nil {
		return err
	}
	return client.Quit()
}
