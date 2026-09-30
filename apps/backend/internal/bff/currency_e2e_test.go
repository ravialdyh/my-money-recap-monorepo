package bff

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestCurrencyChangeEndToEnd (#173): nothing to convert changes both services' code at once; past D57 nothing is
// written; a typed rate answers Home in the new Currency, members following by event; no token converts another's.
func TestCurrencyChangeEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	ledgerURL, ledgerDSN := os.Getenv("LEDGER_TEST_URL"), os.Getenv("LEDGER_TEST_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" || ledgerURL == "" || ledgerDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD, MMR_TEST_MEMBERS_DATABASE_URL, LEDGER_TEST_URL and LEDGER_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, ledgerDB := pool(ctx, t, dsn), pool(ctx, t, ledgerDSN)
	tok := testauth.UserToken(t, kc, pw)
	sub := testauth.Decode(t, tok).Sub
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	keyed := func(req interface{ Header() http.Header }) {
		req.Header().Set("Idempotency-Key", rand.Text())
	}
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, sub); err != nil {
		t.Fatal(err)
	}
	onboard(ctx, t, base, tok)
	done := connect.NewRequest(&moneyrecapv1.BffServiceSaveProfileRequest{OnboardingStep: proto.Int32(setupDone)})
	keyed(done)
	if _, err := client.SaveProfile(ctx, done); err != nil {
		t.Fatal(err)
	}
	id := memberOf(ctx, t, db, sub)
	preview := func(to, rate string) (*moneyrecapv1.CurrencyChangePreview, error) {
		res, err := client.PreviewCurrencyChange(ctx, connect.NewRequest(&moneyrecapv1.BffServicePreviewCurrencyChangeRequest{ToCurrencyCode: to, ManualRate: rate}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Preview, nil
	}
	change := func(to, rate string) (*moneyrecapv1.BffServiceChangeCurrencyResponse, error) {
		msg := &moneyrecapv1.BffServiceChangeCurrencyRequest{ToCurrencyCode: to, Rate: rate}
		if rate != "" {
			msg.Source = moneyrecapv1.RateSource_RATE_SOURCE_MANUAL
		}
		req := connect.NewRequest(msg)
		keyed(req)
		res, err := client.ChangeCurrency(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	membersCode := func() string {
		var code string
		if err := db.QueryRow(ctx, `SELECT currency_code FROM member WHERE id = $1`, id).Scan(&code); err != nil {
			t.Fatal(err)
		}
		return code
	}

	if p, err := preview("GBP", ""); err != nil || !p.NothingToConvert || p.Preview != nil {
		t.Fatalf("a Member with no amounts: preview %v, %v; want nothing_to_convert", p, err)
	}
	direct, err := change("GBP", "")
	if err != nil {
		t.Fatal(err)
	}
	if direct.Home.CurrencyCode != "GBP" || direct.Settings.LedgerSettings.CurrencyCode != "GBP" || membersCode() != "GBP" {
		t.Fatalf("the direct change answered %v with members holding %s, want GBP in both before the answer", direct, membersCode())
	}

	salary := connect.NewRequest(&moneyrecapv1.BffServiceSaveSalaryRequest{Amount: &moneyrecapv1.Money{Amount: "2400.00", CurrencyCode: "GBP"}, Payday: "1"})
	keyed(salary)
	if _, err := client.SaveSalary(ctx, salary); err != nil {
		t.Fatal(err)
	}
	today := time.Now().UTC().Format(dateLayout)
	spend := connect.NewRequest(&moneyrecapv1.BffServiceLogSpendRequest{Amount: &moneyrecapv1.Money{Amount: "12.34", CurrencyCode: "GBP"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: today})
	keyed(spend)
	if _, err := client.LogSpend(ctx, spend); err != nil {
		t.Fatal(err)
	}
	if _, err := preview("IDR", ""); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("no rate: %v, want unavailable, since the stack's currency has no App ID", err)
	}
	if _, err := change("IDR", ""); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("the direct change with amounts: %v, want failed_precondition", err)
	}
	if _, err := preview("IDR", "99999999"); connect.CodeOf(err) != connect.CodeOutOfRange {
		t.Errorf("a preview past 11 digits: %v, want out_of_range", err)
	}
	if _, err := change("IDR", "99999999"); connect.CodeOf(err) != connect.CodeOutOfRange {
		t.Errorf("a conversion past 11 digits: %v, want out_of_range", err)
	}
	conversions(ctx, t, ledgerDB, id, 0)

	p, err := preview("IDR", "20150")
	if err != nil {
		t.Fatal(err)
	}
	if v := p.Preview; v.GetSalaryAfter().GetAmount() != "48360000" || v.GetSpendsBefore().GetAmount() != "12.34" || v.GetSpendsAfter().GetAmount() != "248651" {
		t.Fatalf("preview %v, want £2,400 → Rp48.360.000 and £12.34 → Rp248.651", p)
	}
	converted, err := change("IDR", "20150")
	if err != nil {
		t.Fatal(err)
	}
	if h := converted.Home; h.CurrencyCode != "IDR" || h.CurrencyDigits != 0 || h.Spent.GetAmount() != "248651" || h.Spent.GetCurrencyCode() != "IDR" || converted.Settings.Profile.CurrencyCode != "IDR" {
		t.Fatalf("the conversion answered %v, want Home spent Rp248.651 printed in IDR at once", converted)
	}
	conversions(ctx, t, ledgerDB, id, 1)
	within(t, "members takes IDR from ledger.currency.converted", func() bool { return membersCode() == "IDR" })
	list, err := client.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.BffServiceListSpendsRequest{}))
	if err != nil || list.Msg.Spends.CurrencyCode != "IDR" || list.Msg.Spends.Spends[0].OriginalAmount.GetAmount() != "12.34" {
		t.Errorf("the Spend list: %v, %v; want IDR with the £12.34 kept as typed", list, err)
	}

	other := "00000000-0000-4000-8000-0000000c1730"
	t.Cleanup(func() {
		for _, table := range []string{"member_facts", "salary_setting"} {
			if _, err := ledgerDB.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = $1`, other); err != nil {
				t.Logf("cleanup: %v", err)
			}
		}
	})
	if _, err := ledgerDB.Exec(ctx, `INSERT INTO member_facts (member_id, keycloak_subject, timezone, signed_up_on, currency_code) VALUES ($1, 'someone-else', 'UTC', '2026-09-01', 'GBP')`, other); err != nil {
		t.Fatal(err)
	}
	if _, err := ledgerDB.Exec(ctx, `INSERT INTO salary_setting (member_id, amount, original_amount, original_currency, payday, effective_from) VALUES ($1, 2400, 2400, 'GBP', '1', '2026-09-01')`, other); err != nil {
		t.Fatal(err)
	}
	ledger := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, ledgerURL, testauth.Bearer(tok))
	if _, err := ledger.ApplyCurrencyConversion(ctx, connect.NewRequest(&moneyrecapv1.ApplyCurrencyConversionRequest{
		IdempotencyKey: rand.Text(), MemberId: other, Rate: "20150", Source: moneyrecapv1.RateSource_RATE_SOURCE_MANUAL, ToCurrencyCode: "IDR",
	})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("converting another Member's amounts: %v, want not_found", err)
	}
	var amount string
	if err := ledgerDB.QueryRow(ctx, `SELECT amount::text FROM salary_setting WHERE member_id = $1`, other).Scan(&amount); err != nil || amount != "2400" {
		t.Errorf("the other Member's Salary: %s (%v), want 2400 untouched", amount, err)
	}
	conversions(ctx, t, ledgerDB, other, 0)
}

// conversions checks how many CurrencyConversion rows ledger holds for the Member.
func conversions(ctx context.Context, t *testing.T, ledgerDB *pgxpool.Pool, id string, want int) {
	t.Helper()
	var n int
	if err := ledgerDB.QueryRow(ctx, `SELECT count(*) FROM currency_conversion WHERE member_id = $1`, id).Scan(&n); err != nil || n != want {
		t.Fatalf("%d CurrencyConversion rows (%v), want %d", n, err, want)
	}
}
