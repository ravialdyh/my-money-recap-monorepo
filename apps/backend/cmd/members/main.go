// The members service binary: `members` serves, `members migrate` applies the schema and exits.
package main

import (
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/members"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run is the start-up every database-backed binary shares (internal/serve), given this service's
// name, its schema step and its handler.
func run() error { return serve.Main("members", members.Migrate, handler) }

// handler is MembersService behind the bearer check. It reads no variable of its own.
func handler(_ *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	return moneyrecapv1connect.NewMembersServiceHandler(&members.Service{DB: db}, check)
}
