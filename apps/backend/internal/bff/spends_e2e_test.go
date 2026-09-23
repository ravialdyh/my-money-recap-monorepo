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
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestLogSpendEndToEnd logs Spends through the BFF and ledger with the generated client (#129): today's moves
// Home's Spent, one 40 days back answers this month's Home unchanged, a bad Category is refused. Skips like
// TestEndToEnd.
func TestLogSpendEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD and MMR_TEST_MEMBERS_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	tok := testauth.UserToken(t, kc, pw)
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, testauth.Decode(t, tok).Sub); err != nil {
		t.Fatal(err)
	}
	onboard(ctx, t, base, tok)
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	log := func(day time.Time, category moneyrecapv1.Category) (*moneyrecapv1.BffServiceLogSpendResponse, error) {
		req := connect.NewRequest(&moneyrecapv1.BffServiceLogSpendRequest{
			Amount: &moneyrecapv1.Money{Amount: "25000", CurrencyCode: "IDR"}, Category: category, Date: day.Format(dateLayout), Note: "Kopi susu",
		})
		req.Header().Set("Idempotency-Key", rand.Text())
		res, err := client.LogSpend(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}

	today := time.Now().UTC()
	logged, err := log(today, moneyrecapv1.Category_CATEGORY_SNACKS)
	if err != nil {
		t.Fatal(err)
	}
	if h := logged.Home; h.Spent.Amount != "25000" || h.Spent.CurrencyCode != "IDR" || h.CurrencyDigits != 0 || h.Month != today.Format(monthLayout) {
		t.Fatalf("Home after today's Spend: %v, want Spent 25000 IDR with 0 decimals this month", h)
	}
	if s := logged.Spend; s.Id == "" || s.Note != "Kopi susu" || s.Category != moneyrecapv1.Category_CATEGORY_SNACKS || s.Amount.Amount != "25000" {
		t.Fatalf("Spend: %v, want the one logged, stored", s)
	}
	home, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if home.Msg.Home.Spent.Amount != "25000" {
		t.Fatalf("GetHome Spent %v, want the 25000 just logged", home.Msg.Home.Spent)
	}
	past, err := log(today.AddDate(0, 0, -40), moneyrecapv1.Category_CATEGORY_RIDES)
	if err != nil {
		t.Fatal(err)
	}
	if h := past.Home; h.Month != today.Format(monthLayout) || h.Spent.Amount != "25000" {
		t.Fatalf("Home after a Spend 40 days back: %v, want this month's, Spent still 25000", h)
	}
	if _, err := log(today, moneyrecapv1.Category_CATEGORY_UNSPECIFIED); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a Spend with no Category: %v, want invalid_argument", err)
	}
	t.Logf("logged today and 40 days back; Home Spent %s", past.Home.Spent.Amount)
}
