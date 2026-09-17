// The members service binary: `members` serves, `members migrate` applies the schema and exits.
package main

import (
	"context"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	db, err := pgxpool.New(ctx, serve.Env("MEMBERS_DATABASE_URL", ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := members.Migrate(ctx, db); err != nil {
			log.Fatal(err)
		}
		log.Print("migrations applied")
		return
	}
	if len(os.Args) > 1 {
		log.Fatalf("usage: %s [migrate]", os.Args[0])
	}

	verifier := auth.NewVerifier(ctx, serve.Env("MEMBERS_OIDC_ISSUER", ""), serve.Env("MEMBERS_OIDC_JWKS_URL", ""), serve.Env("MEMBERS_OIDC_AUDIENCE", "bff"))
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewMembersServiceHandler(&members.Service{DB: db}, connect.WithInterceptors(verifier.Interceptor())))
	serve.Run(ctx, "members", serve.Env("MEMBERS_LISTEN", ":8080"), mux, db.Ping)
}
