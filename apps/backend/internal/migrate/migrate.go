// Package migrate is the schema step every service shares (ARCHITECTURE §10): the same image with
// the `migrate` subcommand, run once before the service starts.
package migrate

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run applies every file under files' "migrations" directory not yet recorded in schema_migrations,
// in file-name order (embed.FS lists a directory sorted by name), each in its own transaction. Safe
// to rerun.
func Run(ctx context.Context, db *pgxpool.Pool, files fs.FS) error {
	entries, err := fs.ReadDir(files, "migrations")
	if err != nil {
		return err
	}
	for _, e := range entries {
		sql, err := fs.ReadFile(files, "migrations/"+e.Name())
		if err != nil {
			return err
		}
		err = pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
			if err := lockMigrationLedger(ctx, tx); err != nil {
				return err
			}
			var applied bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE name = $1)`, e.Name()).Scan(&applied); err != nil {
				return err
			}
			if applied {
				return nil
			}
			if _, err := tx.Exec(ctx, string(sql)); err != nil {
				return fmt.Errorf("%s: %w", e.Name(), err)
			}
			_, err := tx.Exec(ctx, `INSERT INTO schema_migrations (name) VALUES ($1)`, e.Name())
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// ledgerLock is the advisory-lock key a migrate run holds for the length of each file's transaction.
// One arbitrary constant serves every service, because the lock is per database and each service has
// its own (ARCHITECTURE A9).
const ledgerLock = 5138619

// lockMigrationLedger serialises two migrate runs racing on one database, so the ledger is created
// once and each file applies once. CREATE TABLE IF NOT EXISTS is not race-safe for concurrent DDL;
// an advisory lock needs no schema to guard it.
func lockMigrationLedger(ctx context.Context, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, ledgerLock); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`)
	return err
}
