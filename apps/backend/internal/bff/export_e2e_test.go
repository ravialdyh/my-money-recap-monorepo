package bff

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestRequestExportEndToEnd (#182): one message in Mailpit to the Member's address with the CSV attached; the same
// key again sends nothing. Two calls, onboarding only when needed: the suite shares 120 a minute (§4.4).
func TestRequestExportEndToEnd(t *testing.T) {
	base, kc, pw, mailpit := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MAILPIT")
	if base == "" || kc == "" || pw == "" || mailpit == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD and MMR_TEST_MAILPIT not all set")
	}
	ctx := context.Background()
	tok := testauth.UserToken(t, kc, pw)
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	query := `to:"` + testauth.Decode(t, tok).Email + `" subject:"Your spends from My Money Recap"`
	before := len(testauth.Search(t, mailpit, query, 0))
	req := connect.NewRequest(&moneyrecapv1.BffServiceRequestExportRequest{})
	req.Header().Set("Idempotency-Key", rand.Text())
	_, err := client.RequestExport(ctx, req)
	if connect.CodeOf(err) == connect.CodeFailedPrecondition {
		onboard(ctx, t, base, tok)
		_, err = client.RequestExport(ctx, req)
	}
	if err != nil {
		t.Fatal(err)
	}
	sent := testauth.Search(t, mailpit, query, before+1)
	if len(sent) != before+1 {
		t.Fatalf("%d export messages after one request, want %d", len(sent), before+1)
	}
	name, csv := testauth.Attachment(t, mailpit, sent[0])
	if header := "date,category,amount,currency,note\r\n"; name != "my-money-recap-spends.csv" || !strings.HasPrefix(string(csv), header) {
		t.Errorf("attachment %q holding %q, want my-money-recap-spends.csv under %q", name, csv, header)
	}
	if _, err := client.RequestExport(ctx, req); err != nil {
		t.Fatal(err)
	}
	if again := testauth.Search(t, mailpit, query, before+2); len(again) != before+1 {
		t.Fatalf("%d export messages after the same key twice, want %d", len(again), before+1)
	}
}
