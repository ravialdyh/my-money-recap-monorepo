package migrate

import (
	"context"
	"crypto/rand"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
)

// A second run of a migration step must change nothing, because Compose runs it before every start
// of the service (ARCHITECTURE §10), and the files must apply in name order.
func TestEveryFileAppliesOnceInNameOrder(t *testing.T) {
	dsn := os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTIFY_TEST_DATABASE_URL not set")
	}
	ctx := t.Context()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)

	table := "probe_" + strings.ToLower(rand.Text())
	files := fstest.MapFS{
		"migrations/0001_" + table + ".sql": {Data: []byte("CREATE TABLE " + table + " (n int)")},
		"migrations/0002_" + table + ".sql": {Data: []byte("INSERT INTO " + table + " (n) VALUES (1)")},
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DROP TABLE IF EXISTS `+table); err != nil {
			t.Logf("cleanup: %v", err)
		}
		if _, err := db.Exec(context.Background(), `DELETE FROM schema_migrations WHERE name LIKE $1`, "%"+table+"%"); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})

	if err := Run(ctx, db, files); err != nil {
		t.Fatal(err)
	}
	if err := Run(ctx, db, files); err != nil {
		t.Fatal(err)
	}
	var rows, applied int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM `+table).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM schema_migrations WHERE name LIKE $1`, "%"+table+"%").Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if rows != 1 || applied != 2 {
		t.Fatalf("two runs left %d rows and %d recorded files, want 1 and 2", rows, applied)
	}
	t.Logf("two runs of two files: %d row, %d recorded", rows, applied)
}

// A file that fails names itself, so a broken migration is not a bare SQL error.
func TestAFailedFileNamesItself(t *testing.T) {
	dsn := os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("NOTIFY_TEST_DATABASE_URL not set")
	}
	db, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	err = Run(t.Context(), db, fstest.MapFS{"migrations/0001_broken.sql": {Data: []byte("NOT SQL")}})
	if err == nil || !strings.Contains(err.Error(), "0001_broken.sql") {
		t.Fatalf("a broken file: %v, want it named", err)
	}
	t.Logf("a broken file: %v", err)
}
