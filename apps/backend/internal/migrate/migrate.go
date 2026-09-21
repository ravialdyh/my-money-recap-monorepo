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
	if _, err := db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
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

// lockMigrationLedger serialises two migrate runs racing on one database, so each file applies once.
func lockMigrationLedger(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `LOCK TABLE schema_migrations IN ACCESS EXCLUSIVE MODE`)
	return err
}
