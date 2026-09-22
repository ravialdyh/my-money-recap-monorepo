package idempotency_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/ledger"
)

// Write runs once per Member and key and replays the stored answer; the key on another request is
// ErrReused; a failed write keeps no key; day-old keys are swept (A14). It uses ledger's table.
func TestWriteRunsOncePerKey(t *testing.T) {
	dsn := os.Getenv("LEDGER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("LEDGER_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := ledger.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	member := "00000000-0000-4000-8000-0000000000e1"
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM idempotency WHERE member_id = $1`, member); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	writes := 0
	run := func(key string, percent int32, fail bool) (*moneyrecapv1.SaveTargetResponse, error) {
		answer := &moneyrecapv1.SaveTargetResponse{}
		req := &moneyrecapv1.SaveTargetRequest{IdempotencyKey: key, MemberId: member, Percent: percent}
		err := idempotency.Write(ctx, db, member, key, req, answer, func(pgx.Tx) error {
			if fail {
				return errors.New("the write failed")
			}
			writes++
			answer.Settings = &moneyrecapv1.LedgerSettings{CurrencyCode: fmt.Sprint(writes)}
			return nil
		})
		return answer, err
	}
	if a, err := run("k1", 20, false); err != nil || writes != 1 || a.Settings.CurrencyCode != "1" {
		t.Fatalf("first: %v, %v, %d writes", a, err, writes)
	}
	if a, err := run("k1", 20, false); err != nil || writes != 1 || a.Settings.CurrencyCode != "1" {
		t.Fatalf("repeat: %v, %v, %d writes; want the stored answer and no second write", a, err, writes)
	}
	if _, err := run("k1", 30, false); !errors.Is(err, idempotency.ErrReused) || writes != 1 {
		t.Fatalf("k1 on another request: %v, %d writes; want ErrReused and no write", err, writes)
	}
	if _, err := run("k2", 20, true); err == nil {
		t.Fatal("a failed write answered no error")
	}
	if a, err := run("k2", 20, false); err != nil || writes != 2 || a.Settings.CurrencyCode != "2" {
		t.Fatalf("k2 after its failed write: %v, %v, %d writes; want it written this time", a, err, writes)
	}
	if _, err := db.Exec(ctx, `INSERT INTO idempotency (member_id, idempotency_key, request_hash, created_at) VALUES ($1, 'old', '\x00', now() - interval '25 hours')`, member); err != nil {
		t.Fatal(err)
	}
	if _, err := run("k3", 20, false); err != nil {
		t.Fatal(err)
	}
	var old int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM idempotency WHERE member_id = $1 AND idempotency_key = 'old'`, member).Scan(&old); err != nil || old != 0 {
		t.Fatalf("%d day-old keys after a write (%v), want the sweep to take it", old, err)
	}
}
