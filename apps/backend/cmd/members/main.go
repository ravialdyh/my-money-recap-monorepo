// The members service binary: `members` serves, `members migrate` applies the schema and exits.
package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/members"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run is the start-up every database-backed binary shares (internal/serve), given this service's
// name, its schema step, its handler and the outbox relay.
func run() error { return serve.Main("members", members.Migrate, handler, relay) }

// handler is MembersService behind the bearer check.
func handler(_ *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	return moneyrecapv1connect.NewMembersServiceHandler(&members.Service{DB: db}, check)
}

// relay publishes the outbox to the broker MEMBERS_AMQP_URL names (ARCHITECTURE §4.3).
func relay(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) { events.Relay(ctx, db, url) }
}
