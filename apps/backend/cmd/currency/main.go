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
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/currency"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var env serve.Env
	issuer, jwks := env.Get("CURRENCY_OIDC_ISSUER", ""), env.Get("CURRENCY_OIDC_JWKS_URL", "")
	audience, redisAddr := env.Get("CURRENCY_OIDC_AUDIENCE", "bff"), env.Get("CURRENCY_REDIS_ADDR", "")
	listen, appID := env.Get("CURRENCY_LISTEN", ":8080"), env.Optional("CURRENCY_OPEN_EXCHANGE_RATES_APP_ID")
	if err := env.Err(); err != nil {
		return err
	}
	cache := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() { _ = cache.Close() }()
	verifier := auth.NewVerifier(ctx, issuer, jwks, audience)
	svc := &currency.Service{AppID: appID, Rates: currency.LatestURL, Cache: cache}
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewCurrencyServiceHandler(svc, connect.WithInterceptors(verifier.Interceptor())))
	return serve.Run(ctx, "currency", listen, mux, nil)
}
