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

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestTargetWarningEndToEnd (#132): notify down at save time still answers Home, without the banner; the warning
// fires once and keeps its first words after more Spends; another Member's warning is not_found. Skips like TestEndToEnd.
func TestTargetWarningEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	membersURL, ledgerURL, notifyURL := os.Getenv("MEMBERS_TEST_URL"), os.Getenv("LEDGER_TEST_URL"), os.Getenv("NOTIFY_TEST_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" || membersURL == "" || ledgerURL == "" || notifyURL == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD, MMR_TEST_MEMBERS_DATABASE_URL, MEMBERS_TEST_URL, LEDGER_TEST_URL and NOTIFY_TEST_URL not all set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	tok := testauth.UserToken(t, kc, pw)
	claims := testauth.Decode(t, tok)
	sub := claims.Sub
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	forward := connect.WithInterceptors(auth.Forward())
	notifyDown := &Service{
		Members: moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, membersURL, forward),
		Ledger:  moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, ledgerURL, forward),
		Notify:  moneyrecapv1connect.NewNotifyServiceClient(http.DefaultClient, "http://127.0.0.1:1", forward),
	}
	signedIn := auth.NewContext(ctx, auth.Claims{Subject: sub, Email: claims.Email, Raw: tok})
	today := time.Now().UTC().Format(dateLayout)
	spend := func(s spender, amount string) *moneyrecapv1.HomeView {
		t.Helper()
		req := connect.NewRequest(&moneyrecapv1.BffServiceLogSpendRequest{
			Amount: &moneyrecapv1.Money{Amount: amount, CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: today,
		})
		req.Header().Set("Idempotency-Key", rand.Text())
		res, err := s.LogSpend(signedIn, req)
		if err != nil {
			t.Fatalf("a Spend of %s: %v", amount, err)
		}
		return res.Msg.Home
	}

	freshMemberOverTheLine(ctx, t, db, client, base, tok, sub)
	if h := spend(client, "400000"); h.TargetWarning != nil {
		t.Fatalf("400000 of a 500000 line: banner %v, want none", h.TargetWarning)
	}
	if h := spend(notifyDown, "150000"); h.TargetWarning != nil || h.Spent.GetAmount() != "550000" {
		t.Fatalf("the crossing Spend with notify down: %v, want Spent 550000 and no banner", h)
	}

	freshMemberOverTheLine(ctx, t, db, client, base, tok, sub)
	spend(client, "400000")
	first := "You've spent Rp550.000 of the Rp500.000 you meant to spend this month."
	if time.Now().UTC().Day() == 1 {
		first = "You've passed your saving line. Rp50.000 over."
	}
	if h := spend(client, "150000"); h.TargetWarning.GetMessage() != first {
		t.Fatalf("the crossing Spend: banner %v, want %q", h.TargetWarning, first)
	}
	if h := spend(client, "100000"); h.TargetWarning.GetMessage() != first || h.Spent.GetAmount() != "650000" {
		t.Fatalf("a Spend after the stamp: %v, want Spent 650000 under the first words", h)
	}
	home, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil || home.Msg.Home.TargetWarning.GetMessage() != first {
		t.Fatalf("GetHome after the stamp: %v, %v, want the first words", home, err)
	}
	t.Logf("fired once with %q, kept after Spent moved to 650000", first)

	notify := moneyrecapv1connect.NewNotifyServiceClient(http.DefaultClient, notifyURL, testauth.Bearer(tok))
	state := func(memberID string) error {
		_, err := notify.GetWarningState(ctx, connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: memberID, Month: today[:7]}))
		return err
	}
	var memberID string
	if err := db.QueryRow(ctx, `SELECT id::text FROM member WHERE keycloak_subject = $1`, sub).Scan(&memberID); err != nil {
		t.Fatal(err)
	}
	if err := state(memberID); err != nil {
		t.Errorf("the Member's own warning at notify: %v", err)
	}
	if err := state("00000000-0000-4000-8000-0000000000ff"); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another Member's warning at notify: %v, want not_found", err)
	}
}

// spender is what the Spends above are logged through: the stack's BFF over HTTP, or one in this process.
type spender interface {
	LogSpend(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceLogSpendRequest]) (*connect.Response[moneyrecapv1.BffServiceLogSpendResponse], error)
}

// freshMemberOverTheLine makes the test Member anew, onboarded in UTC with a 1,000,000 Salary paid on the 1st and a
// 50% Target, so the warning line is 500,000: after the 1st a partial sign-up month warns as before Payday (§8.8).
func freshMemberOverTheLine(ctx context.Context, t *testing.T, db *pgxpool.Pool, client moneyrecapv1connect.BffServiceClient, base, tok, sub string) {
	t.Helper()
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, sub); err != nil {
		t.Fatal(err)
	}
	onboard(ctx, t, base, tok)
	salary := connect.NewRequest(&moneyrecapv1.BffServiceSaveSalaryRequest{
		Amount: &moneyrecapv1.Money{Amount: "1000000", CurrencyCode: "IDR"}, Payday: "1", AppliesFrom: moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_THIS_MONTH,
	})
	salary.Header().Set("Idempotency-Key", rand.Text())
	if _, err := client.SaveSalary(ctx, salary); err != nil {
		t.Fatal(err)
	}
	target := connect.NewRequest(&moneyrecapv1.BffServiceSaveTargetRequest{Percent: 50})
	target.Header().Set("Idempotency-Key", rand.Text())
	if _, err := client.SaveTarget(ctx, target); err != nil {
		t.Fatal(err)
	}
}
