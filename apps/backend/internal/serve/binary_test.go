package serve

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
)

// mainEnv gives Main a name whose variables the test owns and a connection string pgxpool accepts
// without dialling, so the start-up runs as far as the subcommand with no database anywhere.
func mainEnv(t *testing.T, args ...string) {
	t.Helper()
	saved := os.Args
	t.Cleanup(func() { os.Args = saved })
	os.Args = append([]string{"mmrtest"}, args...)
	t.Setenv("MMRTEST_DATABASE_URL", "postgres://probe:probe@127.0.0.1:1/probe")
}

// neverHandler fails the test when Main builds a handler on a path that must never serve.
func neverHandler(t *testing.T) Handler {
	return func(*Env, *pgxpool.Pool, connect.Option) (string, http.Handler) {
		t.Error("Main built a handler on a path that never serves")
		return "/", http.NewServeMux()
	}
}

// neverMigrate fails the test when Main applies a schema on a path that must never migrate.
func neverMigrate(t *testing.T) func(context.Context, *pgxpool.Pool) error {
	return func(context.Context, *pgxpool.Pool) error {
		t.Error("Main applied the schema on a path that never migrates")
		return nil
	}
}

// A binary with no connection string must name the variable and stop, because serving without a
// database and migrating without one are both worse than a start-up that fails (A13).
func TestMainNamesAMissingDatabaseURL(t *testing.T) {
	mainEnv(t, "migrate")
	t.Setenv("MMRTEST_DATABASE_URL", "")
	err := Main("mmrtest", neverMigrate(t), neverHandler(t))
	if err == nil || !strings.Contains(err.Error(), "MMRTEST_DATABASE_URL is required") {
		t.Fatalf("no connection string: %v, want the variable named as required", err)
	}
}

// `<name> migrate` is the only subcommand Compose runs, so anything else is the usage line rather
// than a binary that quietly serves instead (ARCHITECTURE §10).
func TestMainRefusesAnUnknownSubcommand(t *testing.T) {
	mainEnv(t, "serve")
	err := Main("mmrtest", neverMigrate(t), neverHandler(t))
	if err == nil || !strings.Contains(err.Error(), "usage: mmrtest [migrate]") {
		t.Fatalf("an unknown subcommand: %v, want the usage line", err)
	}
}

// The migration step applies this service's schema and exits, because Compose waits on it before it
// starts the service, and it never builds a handler.
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

// A migration that fails is the binary's exit status, so Compose stops the start-up rather than
// running a service against a schema that was never applied.
func TestMainAFailedMigrationIsTheError(t *testing.T) {
	mainEnv(t, "migrate")
	broken := errors.New("0002_probe.sql: syntax error")
	err := Main("mmrtest", func(context.Context, *pgxpool.Pool) error { return broken }, neverHandler(t))
	if !errors.Is(err, broken) {
		t.Fatalf("a failed migration: %v, want %v", err, broken)
	}
}
