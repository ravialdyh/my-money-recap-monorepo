// The members service binary: `members` serves, `members migrate` applies the schema and exits.
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

// run is the start-up every database-backed binary shares (internal/serve), given this service's
// name, its schema step, its handler, its two consumers, the outbox relay and the deletion job.
func run() error {
	return serve.Main("members", members.Migrate, handler, consumer, purges, relay, sweep)
}

// handler is MembersService behind the bearer check. MEMBERS_REALM_URL and MEMBERS_CLIENT_SECRET are members' own
// service account, which deletes accounts at MEMBERS_REALM_ADMIN_URL (ARCHITECTURE §7.1 step 4).
func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	account := &auth.ServiceAccount{Realm: env.Get("MEMBERS_REALM_URL", ""), Client: "members", Secret: env.Get("MEMBERS_CLIENT_SECRET", "")}
	keycloak := members.Keycloak{Admin: env.Get("MEMBERS_REALM_ADMIN_URL", ""), Account: account}
	service = &members.Service{DB: db, Keycloak: keycloak}
	return moneyrecapv1connect.NewMembersServiceHandler(service, check)
}

// service is the one Service: handler builds it before serve.Main builds the workers, so the job shares its key.
var service *members.Service

// consumer takes the Currency ledger converted to from the broker MEMBERS_AMQP_URL names (§4.3, §7.1 "Consumes").
func consumer(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, members.Queue, members.Keys, members.ApplyConversion)
	}
}

// relay publishes the outbox to the broker MEMBERS_AMQP_URL names (ARCHITECTURE §4.3).
func relay(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) { events.Relay(ctx, db, url) }
}

// purges stamps the deletion progress from ledger's and notify's purged events (§7.1 step 5).
func purges(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) {
		events.Consume(ctx, db, url, members.PurgesQueue, members.PurgesKeys, members.ApplyPurged)
	}
}

// sweep is the deletion job, every 10 minutes (§7.1 steps 3, 4 and 6).
func sweep(*serve.Env, *pgxpool.Pool) func(context.Context) { return service.Sweep }
