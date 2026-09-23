package ledger

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// Until onboarding ends every step-3 save replaces the setup Salary from the sign-up month, a new Payday
// included, even a month later; after it, a new Payday waits for next month (owner 2026-09-23, D56).
func TestASetupSaveReplacesTheSetupSalary(t *testing.T) {
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a011"
	ctx := seed(t, db, id, "setup", "Asia/Jakarta", "2026-09-10", "IDR", nil, nil)
	clock := at(t, "2026-09-20", "Asia/Jakarta")
	s := &Service{DB: db, Now: func() time.Time { return clock }}
	this := moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_THIS_MONTH
	salary := func(key, amount, payday string, setup bool) *moneyrecapv1.LedgerSettings {
		t.Helper()
		res, err := s.SaveSalary(ctx, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{
			IdempotencyKey: key, MemberId: id, Amount: idr(amount), Payday: payday, AppliesFrom: this, Setup: setup,
		}))
		if err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		return res.Msg.Settings
	}

	salary("s1", "3000000", "25", true)
	if got := salary("s2", "3500000", "last", true); got.Salary.GetPayday() != "last" || got.Salary.GetEffectiveFrom() != "2026-09" || got.NextSalary != nil {
		t.Fatalf("a setup save with a new Payday: %v, want last in force from 2026-09 and nothing waiting", got)
	}
	clock = at(t, "2026-10-02", "Asia/Jakarta")
	if got := salary("s3", "3600000", "25", true); got.Salary.GetAmount().GetAmount() != "3600000" || got.Salary.GetEffectiveFrom() != "2026-09" {
		t.Fatalf("a setup save the month after sign-up: %v, want 3600000 from 2026-09, the sign-up month", got)
	}
	if got := salary("s4", "3600000", "1", false); got.Salary.GetPayday() != "25" || got.NextSalary.GetPayday() != "1" || got.NextSalary.GetEffectiveFrom() != "2026-11" {
		t.Fatalf("an edit after setup: %v, want the 25th now and the 1st from 2026-11 (D56)", got)
	}
}

// A Member who signed up in July and sets both in September after setup starts them in September, leaving the
// delivered months without; during setup a Target starts in the sign-up month and a re-save replaces it (owner 2026-09-23).
func TestFirstSavesStartThisMonthAfterSetup(t *testing.T) {
	db := testDB(t)
	after := "00000000-0000-4000-8000-00000000a014"
	during := "00000000-0000-4000-8000-00000000a015"
	afterCtx := seed(t, db, after, "after", "Asia/Jakarta", "2026-07-10", "IDR", nil, nil)
	duringCtx := seed(t, db, during, "during", "Asia/Jakarta", "2026-07-10", "IDR", nil, nil)
	s := &Service{DB: db, Now: func() time.Time { return at(t, "2026-09-20", "Asia/Jakarta") }}
	save := func(ctx context.Context, id, key string, setup bool) {
		t.Helper()
		if _, err := s.SaveSalary(ctx, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{
			IdempotencyKey: key, MemberId: id, Amount: idr("3000000"), Payday: "25",
			AppliesFrom: moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_THIS_MONTH, Setup: setup,
		})); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
	}
	target := func(ctx context.Context, id, key string, percent int32, setup bool) *moneyrecapv1.LedgerSettings {
		t.Helper()
		res, err := s.SaveTarget(ctx, connect.NewRequest(&moneyrecapv1.SaveTargetRequest{IdempotencyKey: key, MemberId: id, Percent: percent, Setup: setup}))
		if err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		return res.Msg.Settings
	}

	save(afterCtx, after, "s1", false)
	if got := target(afterCtx, after, "t1", 20, false); got.GetSalary().GetEffectiveFrom() != "2026-09" || got.GetSavingTarget().GetEffectiveFrom() != "2026-09" {
		t.Fatalf("first Salary and Target after setup: %v, want both from 2026-09, not the July sign-up month", got)
	}
	august, err := s.GetMonthSummary(afterCtx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: after, Month: "2026-08"}))
	if err != nil || august.Msg.Summary.Salary != nil || august.Msg.Summary.SavingTarget != nil {
		t.Errorf("August after both saves: %v, %v, want no Salary and no Target", august, err)
	}
	save(duringCtx, during, "s1", true)
	target(duringCtx, during, "t1", 20, true)
	if got := target(duringCtx, during, "t2", 35, true); got.GetSavingTarget().GetPercent() != 35 || got.GetSavingTarget().GetEffectiveFrom() != "2026-07" || got.NextSavingTarget != nil {
		t.Fatalf("a setup Target saved again: %v, want 35%% from 2026-07 and nothing waiting", got)
	}
}

// A Currency change during onboarding relabels the Salary, never converts it: amount and original take the new code,
// rounded half up and kept above zero, and the replica takes it at once, so a save right after is accepted (spec #92).
func TestRelabelCurrency(t *testing.T) {
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a012"
	ctx := seed(t, db, id, "relabel", "Europe/London", "2026-09-10", "GBP", []*moneyrecapv1.SalarySetting{
		{Amount: &moneyrecapv1.Money{Amount: "2400.50"}, OriginalAmount: &moneyrecapv1.Money{Amount: "2400.50", CurrencyCode: "GBP"}, Payday: "25", EffectiveFrom: "2026-09"},
		{Amount: &moneyrecapv1.Money{Amount: "0.40"}, OriginalAmount: &moneyrecapv1.Money{Amount: "0.40", CurrencyCode: "GBP"}, Payday: "25", EffectiveFrom: "2026-08"},
	}, nil)
	s := &Service{DB: db, Now: func() time.Time { return at(t, "2026-09-20", "Europe/London") }}
	relabel := func(key, code string) (*moneyrecapv1.LedgerSettings, error) {
		res, err := s.RelabelCurrency(ctx, connect.NewRequest(&moneyrecapv1.RelabelCurrencyRequest{IdempotencyKey: key, MemberId: id, CurrencyCode: code}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Settings, nil
	}

	got, err := relabel("r1", "IDR")
	if err != nil || got.CurrencyCode != "IDR" || got.Salary.GetAmount().GetAmount() != "2401" || got.Salary.GetOriginalAmount().GetCurrencyCode() != "IDR" || got.Salary.GetOriginalAmount().GetAmount() != "2401" {
		t.Fatalf("GBP 2400.50 relabelled: %v, %v, want IDR 2401 with the original moved too", got, err)
	}
	var tiny, code, replica string
	if err := db.QueryRow(ctx, `SELECT amount::text, original_currency FROM salary_setting WHERE member_id = $1 AND effective_from = '2026-08-01'`, id).Scan(&tiny, &code); err != nil || tiny != "1" || code != "IDR" {
		t.Errorf("GBP 0.40 relabelled: %q %q %v, want 1 IDR, never zero", tiny, code, err)
	}
	if err := db.QueryRow(ctx, `SELECT currency_code FROM member_facts WHERE member_id = $1`, id).Scan(&replica); err != nil || replica != "IDR" {
		t.Errorf("the replica holds %q, %v, want IDR at once", replica, err)
	}
	if _, err := s.SaveSalary(ctx, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{IdempotencyKey: "s1", MemberId: id, Amount: idr("15000000"), Payday: "25", Setup: true})); err != nil {
		t.Errorf("a rupiah save right after: %v, want it accepted", err)
	}
	if again, err := relabel("r1", "IDR"); err != nil || again.Salary.GetAmount().GetAmount() != "2401" {
		t.Errorf("r1 again: %v, %v, want the first answer replayed", again, err)
	}
	if _, err := relabel("r1", "EUR"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("r1 reused for another code: %v, want invalid_argument", err)
	}
	for name, code := range map[string]string{"no currency": "ZZZ", "lower case": "eur", "empty": ""} {
		if _, err := relabel("b-"+name, code); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", name, err)
		}
	}
	if _, err := s.RelabelCurrency(ctx, connect.NewRequest(&moneyrecapv1.RelabelCurrencyRequest{MemberId: id, CurrencyCode: "EUR"})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no key: %v, want invalid_argument", err)
	}
	withMembers(t, s, nil)
	stranger := auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})
	if _, err := s.RelabelCurrency(stranger, connect.NewRequest(&moneyrecapv1.RelabelCurrencyRequest{IdempotencyKey: "x", MemberId: id, CurrencyCode: "EUR"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another Member's token: %v, want not_found", err)
	}
}

// A Member who signs up after this month's Payday and answers step 3 gets the Salary the partial-month fixture
// holds, from the sign-up month: September is partial, with Money In 0 and no Saved (PRD §8.11, D52, spec #92).
func TestAStepThreeAfterPaydayIsTheSignUpMonths(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(fixtureDir, "prd-8.11-signup-27-september.json"))
	if err != nil {
		t.Fatal(err)
	}
	var f moneyrecapv1.LedgerFixture
	if err := protojson.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a013"
	ctx := seed(t, db, id, "partial", f.Timezone, f.SignedUpOn, f.CurrencyCode, nil, nil)
	clock := at(t, f.SignedUpOn, f.Timezone)
	s := &Service{DB: db, Now: func() time.Time { return clock }}
	want := f.Salaries[0]
	res, err := s.SaveSalary(ctx, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{
		IdempotencyKey: "s1", MemberId: id, Amount: want.Amount, Payday: want.Payday,
		AppliesFrom: moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_THIS_MONTH, Setup: true,
	}))
	if err != nil || !proto.Equal(res.Msg.Settings.Salary, want) {
		t.Fatalf("step 3 on %s: %v, %v; want the fixture's Salary %v", f.SignedUpOn, res, err, want)
	}
	clock = at(t, f.AsOf, f.Timezone)
	got, err := s.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: "2026-09"}))
	if err != nil {
		t.Fatal(err)
	}
	if m := got.Msg.Summary; !m.Partial || m.MoneyIn.GetAmount() != "0.00" || m.Saved != nil {
		t.Errorf("September read on %s: %v; want partial, Money In 0.00 and no Saved", f.AsOf, m)
	}
}
