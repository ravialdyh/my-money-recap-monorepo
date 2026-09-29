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
// name, its schema step, its handler and the outbox relay.
func run() error { return serve.Main("members", members.Migrate, handler, relay) }

// handler is MembersService behind the bearer check. MEMBERS_REALM_URL and MEMBERS_CLIENT_SECRET are members' own
// service account, which deletes accounts at MEMBERS_REALM_ADMIN_URL (ARCHITECTURE §7.1 step 4).
func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	account := &auth.ServiceAccount{Realm: env.Get("MEMBERS_REALM_URL", ""), Client: "members", Secret: env.Get("MEMBERS_CLIENT_SECRET", "")}
	keycloak := members.Keycloak{Admin: env.Get("MEMBERS_REALM_ADMIN_URL", ""), Account: account}
	return moneyrecapv1connect.NewMembersServiceHandler(&members.Service{DB: db, Keycloak: keycloak}, check)
}

// relay publishes the outbox to the broker MEMBERS_AMQP_URL names (ARCHITECTURE §4.3).
func relay(env *serve.Env, db *pgxpool.Pool) func(context.Context) {
	url := env.Get("MEMBERS_AMQP_URL", "")
	return func(ctx context.Context) { events.Relay(ctx, db, url) }
}
