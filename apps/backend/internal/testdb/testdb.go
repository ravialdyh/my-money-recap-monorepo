// Package testdb is the tests' scratch schema: a service's migrations in a schema of the test's own, so the stack's
// relays, consumers and jobs never meet a test's rows.
package testdb

import (
	"context"
	"crypto/rand"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Scratch opens the database dsnVar names in a new schema built by migrate and dropped when the test ends. It skips
// the test without dsnVar.
func Scratch(t *testing.T, dsnVar string, migrate func(context.Context, *pgxpool.Pool) error) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " not set")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "scratch_" + strings.ToLower(rand.Text())
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	db, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Logf("cleanup %s: %v", schema, err)
		}
		db.Close()
	})
	if _, err := db.Exec(context.Background(), `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	if err := migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
