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
// name, its schema step, its handler and its consumer.
func run() error { return serve.Main("notify", notify.Migrate, handler, consumer) }

// handler is NotifyService behind the bearer check, which EmailMagicLink alone is exempt from by
// procedure name (§4.2 point 3). NOTIFY_MEMBERS_URL takes the ownership read, with the caller's token.
func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	smtp, from := env.Get("NOTIFY_SMTP_ADDR", ""), env.Get("NOTIFY_MAIL_FROM", "")
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, env.Get("NOTIFY_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward()))
	return moneyrecapv1connect.NewNotifyServiceHandler(&notify.Service{DB: db, Members: members, SMTP: smtp, From: from}, check)
}

// consumer is the warning's event backup: ledger's Spend events on the broker NOTIFY_AMQP_URL names (§4.3, §6.2).
func consumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("NOTIFY_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, notify.Queue, notify.Keys, notify.StampFromEvent)
	}
}
