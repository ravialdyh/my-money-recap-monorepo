package serve

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// Handler builds one service's Connect handler from the open pool and the bearer check. It reads any
// variable of its own from env, which Main reports before it serves.
type Handler func(env *Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler)

// Main is the whole of a binary that owns a database (ARCHITECTURE §10): `<name> migrate` applies
// the schema and exits, `<name>` serves handler on <NAME>_LISTEN with /healthz pinging the pool,
// and any other argument is the usage line.
func Main(name string, migrate func(context.Context, *pgxpool.Pool) error, handler Handler) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	prefix := strings.ToUpper(name)
	var env Env
	dsn := env.Get(prefix+"_DATABASE_URL", "")
	if err := env.Err(); err != nil {
		return err
	}
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if len(os.Args) > 1 {
		if os.Args[1] != "migrate" {
			return fmt.Errorf("usage: %s [migrate]", name)
		}
		if err := migrate(ctx, db); err != nil {
			return err
		}
		log.Print("migrations applied")
		return nil
	}

	issuer, jwks := env.Get(prefix+"_OIDC_ISSUER", ""), env.Get(prefix+"_OIDC_JWKS_URL", "")
	audience, addr := env.Get(prefix+"_OIDC_AUDIENCE", "bff"), env.Get(prefix+"_LISTEN", ":8080")
	verifier := auth.NewVerifier(ctx, issuer, jwks, audience)
	mux := http.NewServeMux()
	mux.Handle(handler(&env, db, connect.WithInterceptors(verifier.Interceptor())))
	if err := env.Err(); err != nil {
		return err
	}
	return Run(ctx, name, addr, mux, db.Ping)
}

// StoreFailed answers a failed store call: the caller's own expired deadline or cancel with its code,
// unlogged; anything else logged and unavailable, so no driver text (a role, a database, an address) reaches
// the caller.
func StoreFailed(ctx context.Context, name string, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return connect.NewError(connect.CodeDeadlineExceeded, ctx.Err())
	}
	if ctx.Err() != nil {
		return connect.NewError(connect.CodeCanceled, ctx.Err())
	}
	log.Printf("%s store: %v", name, err)
	return connect.NewError(connect.CodeUnavailable, fmt.Errorf("the %s store is unavailable", name))
}
