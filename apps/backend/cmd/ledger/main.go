package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/ledger"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error { return serve.Main("ledger", ledger.Migrate, handler, consumer, relay) }

func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, env.Get("LEDGER_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward()))
	return moneyrecapv1connect.NewLedgerServiceHandler(&ledger.Service{DB: db, Members: members}, check)
}

func consumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("LEDGER_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, ledger.Queue, ledger.Keys, ledger.ApplyMemberEvent)
	}
}

func relay(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("LEDGER_AMQP_URL", "")
	return func(ctx context.Context) { events.Relay(ctx, db, url) }
}
