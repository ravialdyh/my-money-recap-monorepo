// The notify service binary: `notify` serves, `notify migrate` applies the schema and exits.
package main

import (
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/notify"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run is the start-up every database-backed binary shares (internal/serve), given this service's
// name, its schema step and its handler.
func run() error { return serve.Main("notify", notify.Migrate, handler) }

// handler is NotifyService behind the bearer check, which EmailMagicLink alone is exempt from by
// procedure name (§4.2 point 3). NOTIFY_MEMBERS_URL takes the ownership read, with the caller's token.
func handler(env *serve.Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler) {
	smtp, from := env.Get("NOTIFY_SMTP_ADDR", ""), env.Get("NOTIFY_MAIL_FROM", "")
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, env.Get("NOTIFY_MEMBERS_URL", ""), connect.WithInterceptors(auth.Forward()))
	return moneyrecapv1connect.NewNotifyServiceHandler(&notify.Service{DB: db, Members: members, SMTP: smtp, From: from}, check)
}
