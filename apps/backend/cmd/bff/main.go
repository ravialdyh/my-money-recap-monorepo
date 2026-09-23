// The BFF binary (ARCHITECTURE §6): the only thing the phone talks to. Owns no database.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"connectrpc.com/connect"
	"github.com/redis/go-redis/v9"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/bff"
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
	issuer, jwks := env.Get("BFF_OIDC_ISSUER", ""), env.Get("BFF_OIDC_JWKS_URL", "")
	audience, membersURL := env.Get("BFF_OIDC_AUDIENCE", "bff"), env.Get("BFF_MEMBERS_URL", "")
	ledgerURL, notifyURL, realm := env.Get("BFF_LEDGER_URL", ""), env.Get("BFF_NOTIFY_URL", ""), env.Get("BFF_REALM_URL", "")
	clientSecret, linkBase := env.Get("BFF_CLIENT_SECRET", ""), env.Get("BFF_MAGIC_LINK_BASE_URL", "")
	redisAddr, listen := env.Get("BFF_REDIS_ADDR", ""), env.Get("BFF_LISTEN", ":8080")
	if err := env.Err(); err != nil {
		return err
	}
	cache := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() { _ = cache.Close() }()
	verifier := auth.NewVerifier(ctx, issuer, jwks, audience)
	svc := &bff.Service{
		Members: moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, membersURL, connect.WithInterceptors(auth.Forward())),
		Ledger:  moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, ledgerURL, connect.WithInterceptors(auth.Forward())),
		Notify:  bff.StubNotify{},
		MagicLink: bff.MagicLink{
			Realm:        realm,
			ClientSecret: clientSecret,
			LinkBase:     linkBase,
			Limit:        bff.RedisLimiter{Client: cache},
			Mail:         moneyrecapv1connect.NewNotifyServiceClient(http.DefaultClient, notifyURL),
		},
	}
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewBffServiceHandler(svc, connect.WithInterceptors(verifier.Interceptor(), bff.MemberLimit(bff.RedisLimiter{Client: cache}), bff.Replay(cache))))
	return serve.Run(ctx, "bff", listen, mux, nil)
}
