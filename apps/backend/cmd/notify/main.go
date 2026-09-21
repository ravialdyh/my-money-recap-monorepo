// The notify service binary: `notify` serves, `notify migrate` applies the schema and exits.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/notify"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run holds main's body so every defer still runs when a start-up step fails.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var env serve.Env
	dsn := env.Get("NOTIFY_DATABASE_URL", "")
	if err := env.Err(); err != nil {
		return err
	}
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	switch {
	case len(os.Args) > 1 && os.Args[1] == "migrate":
		if err := notify.Migrate(ctx, db); err != nil {
			return err
		}
		log.Print("migrations applied")
		return nil
	case len(os.Args) > 1:
		return errors.New("usage: notify [migrate]")
	}

	issuer, jwks := env.Get("NOTIFY_OIDC_ISSUER", ""), env.Get("NOTIFY_OIDC_JWKS_URL", "")
	audience, listen := env.Get("NOTIFY_OIDC_AUDIENCE", "bff"), env.Get("NOTIFY_LISTEN", ":8080")
	smtp, from := env.Get("NOTIFY_SMTP_ADDR", ""), env.Get("NOTIFY_MAIL_FROM", "")
	if err := env.Err(); err != nil {
		return err
	}
	verifier := auth.NewVerifier(ctx, issuer, jwks, audience)
	svc := &notify.Service{DB: db, SMTP: smtp, From: from}
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewNotifyServiceHandler(svc, connect.WithInterceptors(verifier.Interceptor())))
	return serve.Run(ctx, "notify", listen, mux, db.Ping)
}
