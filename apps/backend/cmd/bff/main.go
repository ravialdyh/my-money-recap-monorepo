// The BFF binary (ARCHITECTURE §6): the only thing the phone talks to. Owns no database.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/bff"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	verifier := auth.NewVerifier(ctx, env("BFF_OIDC_ISSUER", ""), env("BFF_OIDC_JWKS_URL", ""), env("BFF_OIDC_AUDIENCE", "bff"))
	svc := &bff.Service{
		Members: moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, env("BFF_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward())),
		Ledger:  bff.StubLedger{},
		Notify:  bff.StubNotify{},
	}
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewBffServiceHandler(svc, connect.WithInterceptors(verifier.Interceptor())))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "ok") })
	srv := &http.Server{Addr: env("BFF_LISTEN", ":8080"), Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	log.Printf("bff listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if def == "" {
		log.Fatalf("%s is required", key)
	}
	return def
}
