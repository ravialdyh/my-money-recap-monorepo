// The members service binary: `members` serves, `members migrate` applies the schema and exits.
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
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/members"
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
	dsn := env.Get("MEMBERS_DATABASE_URL", "")
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
		if err := members.Migrate(ctx, db); err != nil {
			return err
		}
		log.Print("migrations applied")
		return nil
	case len(os.Args) > 1:
		return errors.New("usage: members [migrate]")
	}

	issuer, jwks := env.Get("MEMBERS_OIDC_ISSUER", ""), env.Get("MEMBERS_OIDC_JWKS_URL", "")
	audience, listen := env.Get("MEMBERS_OIDC_AUDIENCE", "bff"), env.Get("MEMBERS_LISTEN", ":8080")
	if err := env.Err(); err != nil {
		return err
	}
	verifier := auth.NewVerifier(ctx, issuer, jwks, audience)
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewMembersServiceHandler(&members.Service{DB: db}, connect.WithInterceptors(verifier.Interceptor())))
	return serve.Run(ctx, "members", listen, mux, db.Ping)
}
