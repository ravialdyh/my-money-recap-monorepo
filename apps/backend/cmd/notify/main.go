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
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/notify"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return serve.Main("notify", notify.Migrate, handler, consumer, membersConsumer, schedules, relay)
}

var service *notify.Service

func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	smtp, from := env.Get("NOTIFY_SMTP_ADDR", ""), env.Get("NOTIFY_MAIL_FROM", "")
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, env.Get("NOTIFY_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward()))
	account := &auth.ServiceAccount{Realm: env.Get("NOTIFY_REALM_URL", ""), Client: "notify", Secret: env.Get("NOTIFY_CLIENT_SECRET", "")}
	ledger := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, env.Get("NOTIFY_LEDGER_URL", ""), connect.WithInterceptors(account.Bearer()))
	service = &notify.Service{DB: db, Members: members, Ledger: ledger, SMTP: smtp, From: from, Apple: apple(env)}
	return moneyrecapv1connect.NewNotifyServiceHandler(service, check)
}

func apple(env *serve.Env) *notify.Apple {
	host, topic := env.Get("NOTIFY_APNS_HOST", "https://api.push.apple.com"), env.Get("NOTIFY_APNS_TOPIC", "com.mymoneyrecap.app")
	key, err := notify.ParseAppleKey(env.Optional("NOTIFY_APPLE_PUSH_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	return &notify.Apple{Host: host, Topic: topic, TeamID: env.Optional("NOTIFY_APPLE_TEAM_ID"), KeyID: env.Optional("NOTIFY_APPLE_PUSH_KEY_ID"), Key: key, Client: http.DefaultClient}
}

func consumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("NOTIFY_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, notify.Queue, notify.Keys, service.StampFromEvent)
	}
}

func membersConsumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("NOTIFY_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, notify.MembersQueue, notify.MembersKeys, notify.ApplyMemberEvent)
	}
}

func schedules(*serve.Env, *pgxpool.Pool) func(context.Context) { return service.Schedule }

func relay(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("NOTIFY_AMQP_URL", "")
	return func(ctx context.Context) { events.Relay(ctx, db, url) }
}
