package bff

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestTargetWarningEndToEnd (#132, #133, #135): notify down still answers Home, bannerless, and the event stamps the
// month; the warning fires once and keeps its first words through more Spends and deletes; another Member's is not_found.
func TestTargetWarningEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	membersURL, ledgerURL, notifyURL := os.Getenv("MEMBERS_TEST_URL"), os.Getenv("LEDGER_TEST_URL"), os.Getenv("NOTIFY_TEST_URL")
	ledgerDSN, notifyDSN := os.Getenv("LEDGER_TEST_DATABASE_URL"), os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" || membersURL == "" || ledgerURL == "" || notifyURL == "" || ledgerDSN == "" || notifyDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD, MMR_TEST_MEMBERS_DATABASE_URL, MEMBERS_TEST_URL, LEDGER_TEST_URL, NOTIFY_TEST_URL, LEDGER_TEST_DATABASE_URL and NOTIFY_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, ledgerDB, notifyDB := pool(ctx, t, dsn), pool(ctx, t, ledgerDSN), pool(ctx, t, notifyDSN)
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
	spend := func(s spender, amount string) *moneyrecapv1.BffServiceLogSpendResponse {
		t.Helper()
		req := connect.NewRequest(&moneyrecapv1.BffServiceLogSpendRequest{
			Amount: &moneyrecapv1.Money{Amount: amount, CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: today,
		})
		req.Header().Set("Idempotency-Key", rand.Text())
		res, err := s.LogSpend(signedIn, req)
		if err != nil {
			t.Fatalf("a Spend of %s: %v", amount, err)
		}
		return res.Msg
	}

	freshMemberOverTheLine(ctx, t, db, client, base, tok, sub)
	if h := spend(client, "400000").Home; h.TargetWarning != nil {
		t.Fatalf("400000 of a 500000 line: banner %v, want none", h.TargetWarning)
	}
	if h := spend(notifyDown, "150000").Home; h.TargetWarning != nil || h.Spent.GetAmount() != "550000" {
		t.Fatalf("the crossing Spend with notify down: %v, want Spent 550000 and no banner", h)
	}
	first := "You've spent Rp550.000 of the Rp500.000 you meant to spend this month."
	if time.Now().UTC().Day() == 1 {
		first = "You've passed your saving line. Rp50.000 over."
	}
	backedUp(ctx, t, ledgerDB, notifyDB, memberOf(ctx, t, db, sub), today[:7], first, func() { spend(notifyDown, "1") })

	freshMemberOverTheLine(ctx, t, db, client, base, tok, sub)
	early := spend(client, "400000").Spend
	crossing := spend(client, "150000")
	if h := crossing.Home; h.TargetWarning.GetMessage() != first {
		t.Fatalf("the crossing Spend: banner %v, want %q", h.TargetWarning, first)
	}
	if h := spend(client, "100000").Home; h.TargetWarning.GetMessage() != first || h.Spent.GetAmount() != "650000" {
		t.Fatalf("a Spend after the stamp: %v, want Spent 650000 under the first words", h)
	}
	for _, id := range []string{early.Id, crossing.Spend.Id} {
		req := connect.NewRequest(&moneyrecapv1.BffServiceDeleteSpendRequest{SpendId: id})
		req.Header().Set("Idempotency-Key", rand.Text())
		if _, err := client.DeleteSpend(ctx, req); err != nil {
			t.Fatal(err)
		}
	}
	home, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil || home.Msg.Home.TargetWarning.GetMessage() != first || home.Msg.Home.Spent.GetAmount() != "100000" {
		t.Fatalf("GetHome after deletes back under the line: %v, %v, want Spent 100000 under the first words", home, err)
	}
	t.Logf("fired once with %q, kept after Spent moved to 650000 and back to 100000", first)

	notify := moneyrecapv1connect.NewNotifyServiceClient(http.DefaultClient, notifyURL, testauth.Bearer(tok))
	state := func(memberID string) error {
		_, err := notify.GetWarningState(ctx, connect.NewRequest(&moneyrecapv1.GetWarningStateRequest{MemberId: memberID, Month: today[:7]}))
		return err
	}
	if err := state(memberOf(ctx, t, db, sub)); err != nil {
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

// backedUp is the event backup (#135): with the direct call failed, the crossing Spend's event stamps the month
// within seconds; delivered again after the flag is cleared, it stamps nothing, proved by the event next stages.
func backedUp(ctx context.Context, t *testing.T, ledgerDB, notifyDB *pgxpool.Pool, memberID, month, want string, next func()) {
	t.Helper()
	kept := func() string {
		var message string
		err := notifyDB.QueryRow(ctx, `SELECT target_warning_message FROM month_flag WHERE member_id = $1 AND year_month = $2`, memberID, month).Scan(&message)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			t.Fatal(err)
		}
		return message
	}
	newest := func() string {
		var id string
		if err := ledgerDB.QueryRow(ctx, `SELECT id::text FROM outbox WHERE member_id = $1 ORDER BY created_at DESC LIMIT 1`, memberID).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	yes := func(db *pgxpool.Pool, q, arg string) func() bool {
		return func() bool {
			var ok bool
			if err := db.QueryRow(ctx, q, arg).Scan(&ok); err != nil {
				t.Fatal(err)
			}
			return ok
		}
	}
	within(t, "the crossing Spend's event stamps the month", func() bool { return kept() == want })
	crossing := newest()
	if _, err := notifyDB.Exec(ctx, `DELETE FROM month_flag WHERE member_id = $1`, memberID); err != nil {
		t.Fatal(err)
	}
	if _, err := ledgerDB.Exec(ctx, `UPDATE outbox SET sent_at = NULL WHERE id = $1`, crossing); err != nil {
		t.Fatal(err)
	}
	within(t, "the relay publishes the crossing event again", yes(ledgerDB, `SELECT sent_at IS NOT NULL FROM outbox WHERE id = $1`, crossing))
	next()
	within(t, "notify processes the event behind the redelivery", yes(notifyDB, `SELECT EXISTS (SELECT 1 FROM processed_events WHERE message_id = $1)`, newest()))
	if got := kept(); got != "" {
		t.Errorf("the crossing event delivered again stamped %q, want nothing", got)
	}
	t.Logf("the event alone stamped %q; delivered again, it changed nothing", want)
}

// within polls ok every 100 ms for up to 15 s.
func within(t *testing.T, what string, ok func() bool) {
	t.Helper()
	for deadline := time.Now().Add(15 * time.Second); !ok(); time.Sleep(100 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("after 15 s: %s", what)
		}
	}
}

func pool(ctx context.Context, t *testing.T, dsn string) *pgxpool.Pool {
	t.Helper()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return db
}

func memberOf(ctx context.Context, t *testing.T, db *pgxpool.Pool, sub string) string {
	t.Helper()
	var id string
	if err := db.QueryRow(ctx, `SELECT id::text FROM member WHERE keycloak_subject = $1`, sub).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
