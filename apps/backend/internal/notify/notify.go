// Package notify is the notify service (ARCHITECTURE §7.3). Slice 3 (#68) sends the magic-link
// email and nothing else: its schema, its SMTP client and its Connect handler.
package notify

import (
	"context"
	"embed"
	"errors"
	"log"
	"net/smtp"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/migrate"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies this service's embedded migrations to the notify database (internal/migrate).
func Migrate(ctx context.Context, db *pgxpool.Pool) error { return migrate.Run(ctx, db, migrations) }

// Service implements NotifyService. Slice 3 serves EmailMagicLink alone, which is the one procedure
// outside the bearer check (auth.PreAuth); every other RPC arrives in slice 7 behind it.
type Service struct {
	moneyrecapv1connect.UnimplementedNotifyServiceHandler
	DB *pgxpool.Pool
	// SMTP is the mail host as host:port: Mailpit on a laptop and in CI, a provider on the VM (§7.5).
	SMTP string
	// From is the envelope sender and the From header.
	From string
}

// EmailMagicLink sends the link once per idempotency key: the row is claimed first and a failed
// send rolls it back, so a repeated key answers success without a second email (A14). An empty key
// would dedupe every later send into silence.
func (s *Service) EmailMagicLink(ctx context.Context, req *connect.Request[moneyrecapv1.EmailMagicLinkRequest]) (*connect.Response[moneyrecapv1.EmailMagicLinkResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
		claimed, err := tx.Exec(ctx, `INSERT INTO magic_link_send (idempotency_key, email) VALUES ($1, $2) ON CONFLICT (idempotency_key) DO NOTHING`, req.Msg.IdempotencyKey, req.Msg.Email)
		if err != nil || claimed.RowsAffected() == 0 {
			return err
		}
		return s.send(req.Msg.Email, req.Msg.Link)
	})
	if err != nil {
		log.Printf("notify magic link: %v", err)
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("the magic-link email was not sent"))
	}
	return connect.NewResponse(&moneyrecapv1.EmailMagicLinkResponse{}), nil
}

// send hands one plain-text message to the SMTP host, its body the 15-minute single-use link of PRD
// D32. Mailpit takes it unauthenticated on a laptop and in CI; the provider the VM needs is slice 7.
func (s *Service) send(to, link string) error {
	msg := "From: " + s.From + "\r\nTo: " + to +
		"\r\nSubject: Your My Money Recap sign-in link\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n\r\n" +
		"Tap this link to sign in. It works for 15 minutes, once.\r\n\r\n" + link + "\r\n"
	return smtp.SendMail(s.SMTP, nil, s.From, []string{to}, []byte(msg))
}
