// The ledger binary, the money service: `ledger` serves, `ledger migrate` applies the schema and exits.
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

// run is the start-up every database-backed binary shares (internal/serve), given this service's
// name, its schema step, its handler, its consumer and its outbox relay.
func run() error { return serve.Main("ledger", ledger.Migrate, handler, consumer, relay) }

// handler is LedgerService behind the bearer check. LEDGER_MEMBERS_URL is where the profile read that
// fills the replica goes, carrying the caller's token (ARCHITECTURE §6.2 step 3).
func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, env.Get("LEDGER_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward()))
	return moneyrecapv1connect.NewLedgerServiceHandler(&ledger.Service{DB: db, Members: members}, check)
}

// consumer keeps the replica fresh from members' events on the broker LEDGER_AMQP_URL names, and purges on
// members.deleted (§4.3, §7.2).
func consumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("LEDGER_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, ledger.Queue, ledger.Keys, ledger.ApplyMemberEvent)
	}
}

// relay publishes the Spend events the outbox holds to the same broker (§4.3).
func relay(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("LEDGER_AMQP_URL", "")
	return func(ctx context.Context) { events.Relay(ctx, db, url) }
}
