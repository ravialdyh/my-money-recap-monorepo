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
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/migrate"
)

//go:embed migrations/*.sql
var migrations embed.FS

const retention = 24 * time.Hour

const sendTimeout = 10 * time.Second

var errAddressRefused = errors.New("the mail host refused the address")

func Migrate(ctx context.Context, db *pgxpool.Pool) error { return migrate.Run(ctx, db, migrations) }

type Service struct {
	moneyrecapv1connect.UnimplementedNotifyServiceHandler
	DB      *pgxpool.Pool
	Members moneyrecapv1connect.MembersServiceClient
	Ledger  moneyrecapv1connect.LedgerServiceClient
	SMTP    string
	From    string
	Apple   *Apple

	pushes sync.WaitGroup
}

func (s *Service) write(ctx context.Context, memberID, key string, req, answer proto.Message, write func(pgx.Tx) error) error {
	err := idempotency.Write(ctx, s.DB, memberID, key, req, answer, write)
	var ce *connect.Error
	switch {
	case err == nil:
		return nil
	case errors.As(err, &ce):
		return ce
	case errors.Is(err, idempotency.ErrReused):
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return storeFailed(ctx, err)
}

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
		return s.send(ctx, req.Msg.Email, "Subject: Your My Money Recap sign-in link\r\n"+
			"Content-Type: text/plain; charset=utf-8\r\n\r\n"+
			"Tap this link to sign in. It works for 15 minutes, once.\r\n\r\n"+req.Msg.Link+"\r\n")
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

func (s *Service) send(ctx context.Context, to, content string) error {
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
	return s.deliver(client, host, to, content)
}

func (s *Service) deliver(client *smtp.Client, host, to, content string) error {
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
	if _, err := body.Write([]byte("From: " + s.From + "\r\nTo: " + to + "\r\n" + content)); err != nil {
		return err
	}
	if err := body.Close(); err != nil {
		return err
	}
	return client.Quit()
}
