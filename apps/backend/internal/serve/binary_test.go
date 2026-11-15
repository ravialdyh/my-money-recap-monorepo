package serve

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mainEnv(t *testing.T, args ...string) {
	t.Helper()
	saved := os.Args
	t.Cleanup(func() { os.Args = saved })
	os.Args = append([]string{"mmrtest"}, args...)
	t.Setenv("MMRTEST_DATABASE_URL", "postgres://probe:probe@127.0.0.1:1/probe")
}

func neverHandler(t *testing.T) Handler {
	return func(*Env, *pgxpool.Pool, connect.Option) (string, http.Handler) {
		t.Error("Main built a handler on a path that never serves")
		return "/", http.NewServeMux()
	}
}

func neverMigrate(t *testing.T) func(context.Context, *pgxpool.Pool) error {
	return func(context.Context, *pgxpool.Pool) error {
		t.Error("Main applied the schema on a path that never migrates")
		return nil
	}
}

func TestMainNamesAMissingDatabaseURL(t *testing.T) {
	mainEnv(t, "migrate")
	t.Setenv("MMRTEST_DATABASE_URL", "")
	err := Main("mmrtest", neverMigrate(t), neverHandler(t))
	if err == nil || !strings.Contains(err.Error(), "MMRTEST_DATABASE_URL is required") {
		t.Fatalf("no connection string: %v, want the variable named as required", err)
	}
}

func TestMainRefusesAnUnknownSubcommand(t *testing.T) {
	mainEnv(t, "serve")
	err := Main("mmrtest", neverMigrate(t), neverHandler(t))
	if err == nil || !strings.Contains(err.Error(), "usage: mmrtest [migrate]") {
		t.Fatalf("an unknown subcommand: %v, want the usage line", err)
	}
}

func TestMainMigrateAppliesTheSchemaAndExits(t *testing.T) {
	mainEnv(t, "migrate")
	applied := false
	migrate := func(context.Context, *pgxpool.Pool) error {
		applied = true
		return nil
	}
	if err := Main("mmrtest", migrate, neverHandler(t)); err != nil {
		t.Fatalf("migrate: %v, want nil", err)
	}
	if !applied {
		t.Fatal("migrate returned without applying the schema")
	}
}

func TestMainAFailedMigrationIsTheError(t *testing.T) {
	mainEnv(t, "migrate")
	broken := errors.New("0002_probe.sql: syntax error")
	err := Main("mmrtest", func(context.Context, *pgxpool.Pool) error { return broken }, neverHandler(t))
	if !errors.Is(err, broken) {
		t.Fatalf("a failed migration: %v, want %v", err, broken)
	}
}

func TestStoreFailedHidesTheDriverText(t *testing.T) {
	expired, stop := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer stop()
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	var logged bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(prev) })
	driver := errors.New("dial tcp 172.18.0.5:5432: connection refused")
	for ctx, want := range map[context.Context]connect.Code{expired: connect.CodeDeadlineExceeded, cancelled: connect.CodeCanceled, context.Background(): connect.CodeUnavailable} {
		if err := StoreFailed(ctx, "ledger", driver); connect.CodeOf(err) != want || strings.Contains(err.Error(), "172.18.0.5") {
			t.Errorf("StoreFailed: %v, want %v without the driver's text", err, want)
		}
	}
	if n := strings.Count(logged.String(), "ledger store: "+driver.Error()); n != 1 {
		t.Errorf("logged %q, want the one store failure under the service's name", logged.String())
	}
}
