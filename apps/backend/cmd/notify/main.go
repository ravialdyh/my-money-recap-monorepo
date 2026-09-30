// The notify service binary: `notify` serves, `notify migrate` applies the schema and exits.
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

// run is the start-up every database-backed binary shares (internal/serve), given this service's
// name, its schema step, its handler and its two consumers.
func run() error { return serve.Main("notify", notify.Migrate, handler, consumer, membersConsumer) }

// service is the one Service: handler builds it before serve.Main builds the consumers, whose loops push through it.
var service *notify.Service

// handler is NotifyService behind the bearer check, EmailMagicLink exempt (§4.2 point 3). NOTIFY_MEMBERS_URL
// takes the ownership read with the caller's token; NOTIFY_LEDGER_URL the schedules' reads with notify's own (§4.2).
func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	smtp, from := env.Get("NOTIFY_SMTP_ADDR", ""), env.Get("NOTIFY_MAIL_FROM", "")
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, env.Get("NOTIFY_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward()))
	account := &auth.ServiceAccount{Realm: env.Get("NOTIFY_REALM_URL", ""), Client: "notify", Secret: env.Get("NOTIFY_CLIENT_SECRET", "")}
	ledger := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, env.Get("NOTIFY_LEDGER_URL", ""), connect.WithInterceptors(account.Bearer()))
	service = &notify.Service{DB: db, Members: members, Ledger: ledger, SMTP: smtp, From: from, Apple: apple(env)}
	return moneyrecapv1connect.NewNotifyServiceHandler(service, check)
}

// apple is the push sender (§7.3): a key that does not parse stops the start-up, and no key at all is a laptop's.
func apple(env *serve.Env) *notify.Apple {
	host, topic := env.Get("NOTIFY_APNS_HOST", "https://api.push.apple.com"), env.Get("NOTIFY_APNS_TOPIC", "com.mymoneyrecap.app")
	key, err := notify.ParseAppleKey(env.Optional("NOTIFY_APPLE_PUSH_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	return &notify.Apple{Host: host, Topic: topic, TeamID: env.Optional("NOTIFY_APPLE_TEAM_ID"), KeyID: env.Optional("NOTIFY_APPLE_PUSH_KEY_ID"), Key: key, Client: http.DefaultClient}
}

// consumer is the warning's event backup: ledger's Spend events on the broker NOTIFY_AMQP_URL names (§4.3, §6.2).
func consumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("NOTIFY_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, notify.Queue, notify.Keys, service.StampFromEvent)
	}
}

// membersConsumer keeps the replica from members' events on the same broker (§7.3 "Consumes").
func membersConsumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("NOTIFY_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, notify.MembersQueue, notify.MembersKeys, notify.ApplyMemberEvent)
	}
}
