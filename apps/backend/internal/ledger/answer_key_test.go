package ledger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// testDB opens the Compose stack's ledger database and applies the schema, skipping unless
// LEDGER_TEST_DATABASE_URL is set (infra/compose/README.md).
func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("LEDGER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("LEDGER_TEST_DATABASE_URL not set")
	}
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

// seed writes one Member's replica row and settings rows straight into the database, owned by subject,
// and removes them when the test ends; it answers a context carrying that subject's claims.
func seed(t *testing.T, db *pgxpool.Pool, id, subject, zone, signedUp, code string, salaries []*moneyrecapv1.SalarySetting, targets []*moneyrecapv1.SavingTargetSetting) context.Context {
	t.Helper()
	ctx := context.Background()
	t.Cleanup(func() {
		for _, table := range []string{"member_facts", "salary_setting", "saving_target_setting", "idempotency"} {
			if _, err := db.Exec(ctx, `DELETE FROM `+table+` WHERE member_id = $1`, id); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
	if _, err := db.Exec(ctx, `INSERT INTO member_facts VALUES ($1, $2, $3, $4, $5)`, id, subject, zone, signedUp, code); err != nil {
		t.Fatal(err)
	}
	for _, s := range salaries {
		if _, err := db.Exec(ctx, `INSERT INTO salary_setting (member_id, amount, original_amount, original_currency, payday, effective_from) VALUES ($1, $2, $3, $4, $5, ($6 || '-01')::date)`,
			id, s.Amount.Amount, s.OriginalAmount.Amount, s.OriginalAmount.CurrencyCode, s.Payday, s.EffectiveFrom); err != nil {
			t.Fatal(err)
		}
	}
	for _, g := range targets {
		if _, err := db.Exec(ctx, `INSERT INTO saving_target_setting (member_id, percent, effective_from) VALUES ($1, $2, ($3 || '-01')::date)`, id, g.Percent, g.EffectiveFrom); err != nil {
			t.Fatal(err)
		}
	}
	return auth.NewContext(ctx, auth.Claims{Subject: subject})
}

// at is noon of a "YYYY-MM-DD" date in zone, so the Member's local today is that date.
func at(t *testing.T, date, zone string) time.Time {
	t.Helper()
	loc, err := time.LoadLocation(zone)
	if err != nil {
		t.Fatal(err)
	}
	d, err := time.ParseInLocation(time.DateOnly, date, loc)
	if err != nil {
		t.Fatal(err)
	}
	return d.Add(12 * time.Hour)
}

// TestTheAnswerKeyWithoutSpends runs every PRD §8 fixture with no Spends, SaveSalary, conversion or Recap
// through the service and its database: rows seeded, the clock at asOf, months and settings compared whole.
func TestTheAnswerKeyWithoutSpends(t *testing.T) {
	db := testDB(t)
	ran := 0
	for i, name := range fixtureNames(t, fixtureDir) {
		raw, err := os.ReadFile(filepath.Join(fixtureDir, name+".json"))
		if err != nil {
			t.Fatal(err)
		}
		var f moneyrecapv1.LedgerFixture
		if err := protojson.Unmarshal(raw, &f); err != nil {
			t.Fatal(err)
		}
		if len(f.Spends) > 0 || f.SaveSalary != nil || f.Convert != nil || len(f.Recaps) > 0 || f.ConversionPreview != nil || f.CrossedLine != nil {
			continue
		}
		ran++
		t.Run(name, func(t *testing.T) {
			id := fmt.Sprintf("00000000-0000-4000-8000-%012d", i)
			ctx := seed(t, db, id, "fixture-"+name, f.Timezone, f.SignedUpOn, f.CurrencyCode, f.Salaries, f.SavingTargets)
			asOf := at(t, f.AsOf, f.Timezone)
			s := &Service{DB: db, Now: func() time.Time { return asOf }}
			for _, want := range f.Months {
				got, err := s.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: want.Month}))
				if err != nil {
					t.Fatal(err)
				}
				if !proto.Equal(got.Msg.Summary, want) {
					t.Errorf("%s:\n got %v\nwant %v", want.Month, got.Msg.Summary, want)
				}
			}
			if f.Settings == nil {
				return
			}
			got, err := s.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: id}))
			if err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(got.Msg.Settings, f.Settings) {
				t.Errorf("settings:\n got %v\nwant %v", got.Msg.Settings, f.Settings)
			}
		})
	}
	if ran == 0 {
		t.Fatal("no fixture without Spends ran")
	}
	t.Logf("%d fixtures without Spends ran as the answer key", ran)
}
