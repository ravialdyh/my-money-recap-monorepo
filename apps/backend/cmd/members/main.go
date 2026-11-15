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
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/members"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return serve.Main("members", members.Migrate, handler, consumer, purges, relay, sweep)
}

func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	account := &auth.ServiceAccount{Realm: env.Get("MEMBERS_REALM_URL", ""), Client: "members", Secret: env.Get("MEMBERS_CLIENT_SECRET", "")}
	keycloak := members.Keycloak{Admin: env.Get("MEMBERS_REALM_ADMIN_URL", ""), Account: account}
	service = &members.Service{DB: db, Keycloak: keycloak}
	return moneyrecapv1connect.NewMembersServiceHandler(service, check)
}

var service *members.Service

func consumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, members.Queue, members.Keys, members.ApplyConversion)
	}
}

func relay(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) { events.Relay(ctx, db, url) }
}

func purges(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, members.PurgesQueue, members.PurgesKeys, members.ApplyPurged)
	}
}

func sweep(*serve.Env, *pgxpool.Pool) func(context.Context) { return service.Sweep }
