package bff

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestSettingsEndToEnd (#170, #171): on a Member with last month's Recap and a stamped warning, a Salary edit after
// Payday with no answer is refused; "from next month" keeps Money In; each edit answers GetHome's Home.
func TestSettingsEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	ledgerDSN := os.Getenv("LEDGER_TEST_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" || ledgerDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD, MMR_TEST_MEMBERS_DATABASE_URL and LEDGER_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, ledgerDB := pool(ctx, t, dsn), pool(ctx, t, ledgerDSN)
	tok := testauth.UserToken(t, kc, pw)
	sub := testauth.Decode(t, tok).Sub
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	keyed := func(req interface{ Header() http.Header }) {
		req.Header().Set("Idempotency-Key", rand.Text())
	}
	spend := func(amount string, on time.Time) {
		t.Helper()
		req := connect.NewRequest(&moneyrecapv1.BffServiceLogSpendRequest{
			Amount: &moneyrecapv1.Money{Amount: amount, CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: on.Format(dateLayout),
		})
		keyed(req)
		if _, err := client.LogSpend(ctx, req); err != nil {
			t.Fatal(err)
		}
	}

	freshMemberOverTheLine(ctx, t, db, client, base, tok, sub)
	today := time.Now().UTC()
	first := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	id, lastFirst := memberOf(ctx, t, db, sub), first.AddDate(0, -1, 0)
	if _, err := db.Exec(ctx, `UPDATE member SET signed_up_on = $2 WHERE id = $1`, id, lastFirst); err != nil {
		t.Fatal(err)
	}
	if tag, err := ledgerDB.Exec(ctx, `UPDATE member_facts SET signed_up_on = $2, facts_at = now() WHERE member_id = $1`, id, lastFirst); err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("moving ledger's sign-up date: %v, %d rows", err, tag.RowsAffected())
	}
	spend("25000", first.AddDate(0, 0, -1))
	spend("550000", today)

	sameAsGetHome := func(what string, h *moneyrecapv1.HomeView) {
		t.Helper()
		if len(h.RecapMonths) != 1 || h.RecapMonths[0].Month != lastFirst.Format(monthLayout) || h.TargetWarning == nil || h.EarliestSpendDate == "" {
			t.Fatalf("%s answered Home %v, want last month's Recap, the kept banner and the earliest Spend date", what, h)
		}
		home, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
		if err != nil || !proto.Equal(home.Msg.Home, h) {
			t.Fatalf("%s answered Home %v, GetHome %v (%v); want the same", what, h, home, err)
		}
	}
	before, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	edit := func(applies moneyrecapv1.SalaryAppliesFrom) (*connect.Response[moneyrecapv1.BffServiceSaveSalaryResponse], error) {
		salary := connect.NewRequest(&moneyrecapv1.BffServiceSaveSalaryRequest{
			Amount: &moneyrecapv1.Money{Amount: "2000000", CurrencyCode: "IDR"}, Payday: "1", AppliesFrom: applies,
		})
		keyed(salary)
		return client.SaveSalary(ctx, salary)
	}
	if _, err := edit(moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_UNSPECIFIED); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("an edit on or after the 1st with no D56 answer: %v, want failed_precondition", err)
	}
	saved, err := edit(moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_NEXT_MONTH)
	if err != nil {
		t.Fatal(err)
	}
	sameAsGetHome("SaveSalary", saved.Msg.Home)
	if got, want := saved.Msg.Home.MoneyIn.GetAmount(), before.Msg.Home.MoneyIn.GetAmount(); got != want || want != "1000000" {
		t.Fatalf("Money In after a raise from next month: %s, want this month's %s", got, want)
	}
	target := connect.NewRequest(&moneyrecapv1.BffServiceSaveTargetRequest{Percent: 40})
	keyed(target)
	aimed, err := client.SaveTarget(ctx, target)
	if err != nil {
		t.Fatal(err)
	}
	sameAsGetHome("SaveTarget", aimed.Msg.Home)

	res, err := client.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetSettingsRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	v, next := res.Msg.Settings, first.AddDate(0, 1, 0).Format(monthLayout)
	if l := v.LedgerSettings; v.Profile.DisplayName != "Test" || v.Profile.CurrencyCode != "IDR" || v.Profile.CurrencyDigits != 0 || v.NotificationPrefs != nil ||
		l.Salary.GetAmount().GetAmount() != "1000000" || l.NextSalary.GetAmount().GetAmount() != "2000000" || l.NextSalary.GetEffectiveFrom() != next ||
		l.SavingTarget.GetPercent() != 50 || l.NextSavingTarget.GetPercent() != 40 || l.NextSavingTarget.GetEffectiveFrom() != next {
		t.Fatalf("GetSettings: %v, want Test in IDR, 1000000 now and 2000000 from %s, 50%% now and 40%% from %s, no preferences", v, next, next)
	}
	t.Logf("Settings read after both edits: Salary %s then %s, Target %d%% then %d%%", v.LedgerSettings.Salary.GetAmount().GetAmount(),
		v.LedgerSettings.NextSalary.GetAmount().GetAmount(), v.LedgerSettings.SavingTarget.GetPercent(), v.LedgerSettings.NextSavingTarget.GetPercent())
}
