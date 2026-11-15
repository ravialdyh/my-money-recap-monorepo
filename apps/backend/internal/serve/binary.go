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

type Handler func(env *Env, db *pgxpool.Pool, check connect.Option) (string, http.Handler)

type Worker func(env *Env, db *pgxpool.Pool) func(ctx context.Context)

func Main(name string, migrate func(context.Context, *pgxpool.Pool) error, handler Handler, workers ...Worker) error {
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
	loops := make([]func(context.Context), 0, len(workers))
	for _, w := range workers {
		loops = append(loops, w(&env, db))
	}
	if err := env.Err(); err != nil {
		return err
	}
	for _, loop := range loops {
		go loop(ctx)
	}
	return Run(ctx, name, addr, mux, db.Ping)
}

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
