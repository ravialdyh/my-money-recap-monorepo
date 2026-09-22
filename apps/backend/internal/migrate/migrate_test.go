package migrate

import (
	"context"
	"crypto/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"
)

// probeFiles is one file that creates a table and one that fills it: enough to prove name order and
// a rerun that changes nothing, and nothing a service would recognise.
var probeFiles = fstest.MapFS{
	"migrations/0001_probe.sql": {Data: []byte("CREATE TABLE probe (n int)")},
	"migrations/0002_probe.sql": {Data: []byte("INSERT INTO probe (n) VALUES (1)")},
}

// probePool takes whichever service database the environment offers and points search_path at a
// throwaway schema, dropped when the test ends, so the shared runner's own tests leave neither a
// ledger row nor a probe table in a service's schema.
func probePool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("NOTIFY_TEST_DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("neither MMR_TEST_MEMBERS_DATABASE_URL nor NOTIFY_TEST_DATABASE_URL is set")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "migrate_probe_" + strings.ToLower(rand.Text())
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	cfg.MaxConns = 8
	db, err := pgxpool.NewWithConfig(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(t.Context(), `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Logf("cleanup: %v", err)
		}
		db.Close()
	})
	return db
}

// counts is the probe table's rows and the ledger's, the two numbers a second run must not change.
func counts(t *testing.T, db *pgxpool.Pool) (rows, applied int) {
	t.Helper()
	if err := db.QueryRow(t.Context(), `SELECT count(*) FROM probe`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(t.Context(), `SELECT count(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	return rows, applied
}

// firstRuns calls Run n times at once on connections the pool has already opened, so every caller
// reaches the ledger together; a cold pool opens them one by one and the race mostly does not happen.
func firstRuns(t *testing.T, db *pgxpool.Pool, n int) []error {
	t.Helper()
	warm := make([]*pgxpool.Conn, n)
	for i := range warm {
		c, err := db.Acquire(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		warm[i] = c
	}
	for _, c := range warm {
		c.Release()
	}
	errs := make([]error, n)
	var gate, done sync.WaitGroup
	gate.Add(1)
	done.Add(n)
	for i := range errs {
		go func() {
			defer done.Done()
			gate.Wait()
			errs[i] = Run(t.Context(), db, probeFiles)
		}()
	}
	gate.Done()
	done.Wait()
	return errs
}

// A second run of a migration step must change nothing, because Compose runs it before every start
// of the service (ARCHITECTURE §10), and the files must apply in name order.
func TestEveryFileAppliesOnceInNameOrder(t *testing.T) {
	db := probePool(t)
	if err := Run(t.Context(), db, probeFiles); err != nil {
		t.Fatal(err)
	}
	if err := Run(t.Context(), db, probeFiles); err != nil {
		t.Fatal(err)
	}
	rows, applied := counts(t, db)
	if rows != 1 || applied != 2 {
		t.Fatalf("two runs left %d rows and %d recorded files, want 1 and 2", rows, applied)
	}
	t.Logf("two runs of two files: %d row, %d recorded", rows, applied)
}

// Six first runs racing on an empty database must all succeed, because two services call this runner
// and Compose starts both migration steps at once: CREATE TABLE IF NOT EXISTS is not race-safe, so
// the ledger is created under the advisory lock.
func TestConcurrentFirstRunsDoNotRace(t *testing.T) {
	db := probePool(t)
	for round := range 3 {
		for _, err := range firstRuns(t, db, 6) {
			if err != nil {
				t.Fatalf("round %d: a concurrent first run: %v", round, err)
			}
		}
		rows, applied := counts(t, db)
		if rows != 1 || applied != 2 {
			t.Fatalf("round %d left %d rows and %d recorded files, want 1 and 2", round, rows, applied)
		}
		if _, err := db.Exec(t.Context(), `DROP TABLE probe, schema_migrations`); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("three rounds of six concurrent first runs: no collision, one row and two recorded files each")
}

// A file that fails names itself, so a broken migration is not a bare SQL error.
func TestAFailedFileNamesItself(t *testing.T) {
	db := probePool(t)
	err := Run(t.Context(), db, fstest.MapFS{"migrations/0001_broken.sql": {Data: []byte("NOT SQL")}})
	if err == nil || !strings.Contains(err.Error(), "0001_broken.sql") {
		t.Fatalf("a broken file: %v, want it named", err)
	}
	t.Logf("a broken file: %v", err)
}
