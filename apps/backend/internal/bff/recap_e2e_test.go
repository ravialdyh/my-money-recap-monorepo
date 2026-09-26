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

// TestGetRecapEndToEnd (#152): the current month is the preview; with sign-up moved back, last month's Spend answers
// under the new name, an earlier month is not_found, a later one invalid_argument, and ledger refuses another Member.
func TestGetRecapEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	ledgerURL, ledgerDSN := os.Getenv("LEDGER_TEST_URL"), os.Getenv("LEDGER_TEST_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" || ledgerURL == "" || ledgerDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD, MMR_TEST_MEMBERS_DATABASE_URL, LEDGER_TEST_URL and LEDGER_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, ledgerDB := pool(ctx, t, dsn), pool(ctx, t, ledgerDSN)
	tok := testauth.UserToken(t, kc, pw)
	sub := testauth.Decode(t, tok).Sub
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, sub); err != nil {
		t.Fatal(err)
	}
	onboard(ctx, t, base, tok)
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	recap := func(month string) (*moneyrecapv1.RecapView, error) {
		res, err := client.GetRecap(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetRecapRequest{Month: month}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Recap, nil
	}
	send := func(req interface{ Header() http.Header }) {
		req.Header().Set("Idempotency-Key", rand.Text())
	}

	today := time.Now().UTC()
	first := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
	now, err := recap(today.Format(monthLayout))
	if err != nil || now.Recap.GetSituation() != moneyrecapv1.PartialSituation_PARTIAL_SITUATION_CURRENT_MONTH || now.DisplayName != "Test" || now.CurrencyCode != "IDR" {
		t.Fatalf("the current month: %v, %v, want the preview for Test in IDR", now, err)
	}

	id, lastFirst := memberOf(ctx, t, db, sub), first.AddDate(0, -1, 0)
	if _, err := db.Exec(ctx, `UPDATE member SET signed_up_on = $2 WHERE id = $1`, id, lastFirst); err != nil {
		t.Fatal(err)
	}
	if tag, err := ledgerDB.Exec(ctx, `UPDATE member_facts SET signed_up_on = $2, facts_at = now() WHERE member_id = $1`, id, lastFirst); err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("moving ledger's sign-up date: %v, %d rows", err, tag.RowsAffected())
	}
	spend := connect.NewRequest(&moneyrecapv1.BffServiceLogSpendRequest{
		Amount: &moneyrecapv1.Money{Amount: "25000", CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: first.AddDate(0, 0, -1).Format(dateLayout),
	})
	send(spend)
	if _, err := client.LogSpend(ctx, spend); err != nil {
		t.Fatal(err)
	}
	rename := connect.NewRequest(&moneyrecapv1.BffServiceSaveProfileRequest{DisplayName: proto.String("Tess"), AvatarKey: proto.String("owl"), OnboardingStep: proto.Int32(2)})
	send(rename)
	if _, err := client.SaveProfile(ctx, rename); err != nil {
		t.Fatal(err)
	}

	last, err := recap(lastFirst.Format(monthLayout))
	if r := last.GetRecap(); err != nil || r.Situation != moneyrecapv1.PartialSituation_PARTIAL_SITUATION_NO_MONEY_IN || r.SpendCount != 1 ||
		r.Spent.GetAmount() != "25000" || r.Streak.GetWindowStart() != lastFirst.Format(dateLayout) || last.DisplayName != "Tess" {
		t.Fatalf("last month: %v, %v, want its one Spend as a Partial Recap under the new name", last, err)
	}
	if _, err := recap(lastFirst.AddDate(0, -1, 0).Format(monthLayout)); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("the month before sign-up: %v, want not_found", err)
	}
	if _, err := recap(first.AddDate(0, 1, 0).Format(monthLayout)); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("next month: %v, want invalid_argument", err)
	}
	ledger := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, ledgerURL, testauth.Bearer(tok))
	if _, err := ledger.GetRecap(ctx, connect.NewRequest(&moneyrecapv1.GetRecapRequest{MemberId: "00000000-0000-4000-8000-0000000000ff", Month: today.Format(monthLayout)})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another Member's Recap at ledger: %v, want not_found", err)
	}
	t.Logf("last month's Recap for %s: %d Spend, Spent %s", last.DisplayName, last.Recap.SpendCount, last.Recap.Spent.GetAmount())
}
