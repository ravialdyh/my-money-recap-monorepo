// The BFF binary (ARCHITECTURE §6): the only thing the phone talks to. Owns no database.
package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"connectrpc.com/connect"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/bff"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	verifier := auth.NewVerifier(ctx, serve.Env("BFF_OIDC_ISSUER", ""), serve.Env("BFF_OIDC_JWKS_URL", ""), serve.Env("BFF_OIDC_AUDIENCE", "bff"))
	svc := &bff.Service{
		Members: moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, serve.Env("BFF_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward())),
		Ledger:  bff.StubLedger{},
		Notify:  bff.StubNotify{},
	}
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewBffServiceHandler(svc, connect.WithInterceptors(verifier.Interceptor())))
	serve.Run(ctx, "bff", serve.Env("BFF_LISTEN", ":8080"), mux, nil)
}
