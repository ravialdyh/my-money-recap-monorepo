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

// TestLogSpendEndToEnd logs, lists, edits and deletes a Spend through the BFF and ledger (#129, #131, #133); a Spend
// before the sign-up month (D72) or with no Category is refused.
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

	list := func(month string) *moneyrecapv1.SpendListView {
		t.Helper()
		res, err := client.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.BffServiceListSpendsRequest{Month: month}))
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg.Spends
	}

	today := time.Now().UTC()
	if v := list(""); v.Month != today.Format(monthLayout) || len(v.Spends) != 0 || v.Spent.Amount != "0" || v.StreakDay != 1 || v.CurrencyDigits != 0 {
		t.Fatalf("the empty month: %v, want this month at 0 on streak day 1, IDR's 0 decimals", v)
	}
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
	v := list("")
	if len(v.Spends) != 1 || v.Spends[0].Id != logged.Spend.Id || len(v.Categories) != 1 || v.Categories[0].Percent != 100 ||
		v.Categories[0].SpendCount != 1 || v.Spent.Amount != "25000" || len(v.Months) != 1 || v.Months[0] != v.Month {
		t.Fatalf("this month's list: %v, want the Spend as Snacks at 100%% and only this month to switch to", v)
	}
	first := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, -1, 0).Format(monthLayout)
	if v := list(last); v.Month != last || len(v.Spends)+len(v.Categories) != 0 || v.Spent.Amount != "0" || v.StreakDay != 0 {
		t.Fatalf("last month, before sign-up: %v, want nothing and no Streak", v)
	}
	if home.Msg.Home.EarliestSpendDate != first.Format(dateLayout) {
		t.Fatalf("earliest date %q, want the 1st of the sign-up month (D72)", home.Msg.Home.EarliestSpendDate)
	}
	if _, err := log(first.AddDate(0, 0, -1), moneyrecapv1.Category_CATEGORY_RIDES); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a Spend before the sign-up month: %v, want invalid_argument (D72)", err)
	}
	if _, err := log(today, moneyrecapv1.Category_CATEGORY_UNSPECIFIED); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a Spend with no Category: %v, want invalid_argument", err)
	}
	editAndDelete(ctx, t, client, logged.Spend.Id, first)
}

// editAndDelete edits the Spend with id to 40000 of Rides on the 1st, which D72 allows the test Member who signs up
// this month, refuses a move to last month, then deletes it: Home and the list follow, and it is gone for good (#133).
func editAndDelete(ctx context.Context, t *testing.T, client moneyrecapv1connect.BffServiceClient, id string, first time.Time) {
	t.Helper()
	edit := func(day time.Time) (*moneyrecapv1.BffServiceEditSpendResponse, error) {
		req := connect.NewRequest(&moneyrecapv1.BffServiceEditSpendRequest{
			SpendId: id, Amount: &moneyrecapv1.Money{Amount: "40000", CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_RIDES, Date: day.Format(dateLayout), Note: "Ojek",
		})
		req.Header().Set("Idempotency-Key", rand.Text())
		res, err := client.EditSpend(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	del := func() error {
		req := connect.NewRequest(&moneyrecapv1.BffServiceDeleteSpendRequest{SpendId: id})
		req.Header().Set("Idempotency-Key", rand.Text())
		_, err := client.DeleteSpend(ctx, req)
		return err
	}
	edited, err := edit(first)
	if err != nil || edited.Home.Spent.Amount != "40000" || edited.Spend.Id != id || edited.Spend.Date != first.Format(dateLayout) ||
		edited.Spend.Category != moneyrecapv1.Category_CATEGORY_RIDES || edited.Spend.Note != "Ojek" {
		t.Fatalf("the Spend edited to 40000 of Rides on the 1st: %v, %v", edited, err)
	}
	if _, err := edit(first.AddDate(0, 0, -1)); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("the Spend moved before the sign-up month: %v, want invalid_argument (D72)", err)
	}
	if err := del(); err != nil {
		t.Fatal(err)
	}
	list, err := client.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.BffServiceListSpendsRequest{}))
	if err != nil || len(list.Msg.Spends.Spends) != 0 || list.Msg.Spends.Spent.Amount != "0" {
		t.Fatalf("the list after the delete: %v, %v, want it empty", list, err)
	}
	if err := del(); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("the deleted Spend deleted again: %v, want not_found", err)
	}
}
