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
		for _, table := range []string{"member_facts", "salary_setting", "saving_target_setting", "spend", "idempotency"} {
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

// seedSpends writes Spend rows straight into the database; seed's cleanup removes them.
func seedSpends(t *testing.T, db *pgxpool.Pool, id string, spends []*moneyrecapv1.Spend) {
	t.Helper()
	for _, s := range spends {
		if _, err := db.Exec(context.Background(), `INSERT INTO spend (member_id, amount, original_amount, original_currency, category, spent_on, note) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id, s.Amount.Amount, s.OriginalAmount.Amount, s.OriginalAmount.CurrencyCode, int16(s.Category), s.Date, s.Note); err != nil {
			t.Fatal(err)
		}
	}
}

// leftOut names why a fixture does not run yet, or nothing when it runs.
func leftOut(f *moneyrecapv1.LedgerFixture) string {
	if f.Convert != nil || f.ConversionPreview != nil {
		return "the currency change waits for slice 7"
	}
	return ""
}

// TestTheAnswerKey runs every PRD §8 fixture through the service and its database: rows seeded, the last Spend
// logged on its own date where crossed_line is asserted, the Salary saved, then months, Recaps, settings and rows compared.
func TestTheAnswerKey(t *testing.T) {
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
		if why := leftOut(&f); why != "" {
			t.Logf("left out: %s: %s", name, why)
			continue
		}
		ran++
		t.Run(name, func(t *testing.T) { runFixture(t, db, fmt.Sprintf("00000000-0000-4000-8000-%012d", i), name, &f) })
	}
	if ran != 23 {
		t.Fatalf("%d fixtures ran, want 23", ran)
	}
}

func runFixture(t *testing.T, db *pgxpool.Pool, id, name string, f *moneyrecapv1.LedgerFixture) {
	ctx := seed(t, db, id, "fixture-"+name, f.Timezone, f.SignedUpOn, f.CurrencyCode, f.Salaries, f.SavingTargets)
	seeded := f.Spends
	if f.CrossedLine != nil {
		seeded = f.Spends[:len(f.Spends)-1]
	}
	seedSpends(t, db, id, seeded)
	clock := at(t, f.AsOf, f.Timezone)
	s := &Service{DB: db, Now: func() time.Time { return clock }}
	if f.CrossedLine != nil {
		last := f.Spends[len(f.Spends)-1]
		clock = at(t, last.Date, f.Timezone)
		got, err := s.LogSpend(ctx, connect.NewRequest(&moneyrecapv1.LogSpendRequest{
			IdempotencyKey: "log", MemberId: id, Amount: last.Amount, Category: last.Category, Date: last.Date, Note: last.Note,
		}))
		if err != nil {
			t.Fatal(err)
		}
		if got.Msg.CrossedLine != *f.CrossedLine {
			t.Errorf("crossed_line %v, want %v", got.Msg.CrossedLine, *f.CrossedLine)
		}
		logged := got.Msg.Spend
		logged.Id, logged.CreatedAt, logged.UpdatedAt = last.Id, nil, nil
		if !proto.Equal(logged, last) {
			t.Errorf("logged:\n got %v\nwant %v", logged, last)
		}
		clock = at(t, f.AsOf, f.Timezone)
	}
	if f.SaveSalary != nil {
		f.SaveSalary.IdempotencyKey, f.SaveSalary.MemberId = "salary", id
		if _, err := s.SaveSalary(ctx, connect.NewRequest(f.SaveSalary)); err != nil {
			t.Fatal(err)
		}
	}
	for _, want := range f.Months {
		got, err := s.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: want.Month}))
		if err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(got.Msg.Summary, want) {
			t.Errorf("%s:\n got %v\nwant %v", want.Month, got.Msg.Summary, want)
		}
	}
	for _, want := range f.Recaps {
		got, err := s.GetRecap(ctx, connect.NewRequest(&moneyrecapv1.GetRecapRequest{MemberId: id, Month: want.Month}))
		if err != nil {
			t.Fatal(err)
		}
		if splurge := got.Msg.Recap.Splurge; splurge != nil {
			splurge.Id, splurge.CreatedAt, splurge.UpdatedAt = want.Splurge.GetId(), nil, nil
		}
		if !proto.Equal(got.Msg.Recap, want) {
			t.Errorf("recap %s:\n got %v\nwant %v", want.Month, got.Msg.Recap, want)
		}
	}
	if f.Settings != nil {
		got, err := s.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: id}))
		if err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(got.Msg.Settings, f.Settings) {
			t.Errorf("settings:\n got %v\nwant %v", got.Msg.Settings, f.Settings)
		}
	}
	if len(f.SalariesAfter) > 0 {
		rows, err := salaries(ctx, db, id)
		if err != nil {
			t.Fatal(err)
		}
		c, _ := currencyOf(f.CurrencyCode)
		got := &moneyrecapv1.LedgerFixture{}
		for _, r := range rows {
			got.SalariesAfter = append(got.SalariesAfter, ledgerOf{cash: c}.salarySetting(r))
		}
		if !proto.Equal(got, &moneyrecapv1.LedgerFixture{SalariesAfter: f.SalariesAfter}) {
			t.Errorf("salary rows:\n got %v\nwant %v", got.SalariesAfter, f.SalariesAfter)
		}
	}
}
