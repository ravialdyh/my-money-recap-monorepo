package bff

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"uuid"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestMagicLinkEndToEnd drives the magic-link chain at the BFF's HTTP boundary against the real
// Compose stack (#69, spec #61). It skips unless the four variables it reads are all
// set; infra/compose/README.md lists them.
func TestMagicLinkEndToEnd(t *testing.T) {
	base, kc := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK")
	mailpit, dsn := os.Getenv("MMR_TEST_MAILPIT"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	if base == "" || kc == "" || mailpit == "" || dsn == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_MAILPIT and MMR_TEST_MEMBERS_DATABASE_URL not all set")
	}
	ctx := t.Context()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	anonymous := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(""))

	t.Run("a fresh address signs a Member in, once", func(t *testing.T) {
		address := freshAddress(t)
		forgetMember(t, db, address)
		if _, err := anonymous.RequestMagicLink(ctx, request(address, "")); err != nil {
			t.Fatalf("RequestMagicLink on a client with an empty bearer: %v", err)
		}
		loginToken := loginTokenFromMail(t, mailpit, address)
		t.Logf("%s → %s", address, loginToken)

		access, refused := redeem(t, kc, loginToken)
		if access == "" {
			t.Fatalf("redeeming on the ios client: %s", refused)
		}
		second, refusal := redeem(t, kc, loginToken)
		if second != "" {
			t.Errorf("the second redemption issued another token, want a refusal")
		}
		t.Logf("second redemption of the same token → %s", refusal)

		client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(access))
		if _, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
			t.Fatalf("GetHome for the new Member: %v, want failed_precondition until steps 1 and 2 are answered", err)
		}
		onboard(ctx, t, base, access)
		home, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
		if err != nil {
			t.Fatal(err)
		}
		assertZeroHome(t, home.Msg.Home, "Test", "IDR")
		if rows := memberRows(t, db, address); rows != 1 {
			t.Fatalf("%d Member rows for %s, want 1", rows, address)
		}
		t.Logf("GetHome on the magic-link token → failed_precondition, then after steps 1 and 2 a HomeView of zeros and one Member row")
	})

	t.Run("the sixth request in the hour is refused", func(t *testing.T) {
		address := freshAddress(t)
		for i := range limitPerWindow {
			if _, err := anonymous.RequestMagicLink(ctx, request(address, "")); err != nil {
				t.Fatalf("request %d for %s: %v", i+1, address, err)
			}
		}
		_, err := anonymous.RequestMagicLink(ctx, request(address, ""))
		if connect.CodeOf(err) != connect.CodeResourceExhausted {
			t.Fatalf("request %d for %s: %v, want resource_exhausted", limitPerWindow+1, address, err)
		}
		t.Logf("request %d for %s → %v", limitPerWindow+1, address, err)
	})

	t.Run("one idempotency key sends one email", func(t *testing.T) {
		address, key := freshAddress(t), uuid.NewV4().String()
		for i := range 2 {
			if _, err := anonymous.RequestMagicLink(ctx, request(address, key)); err != nil {
				t.Fatalf("request %d carrying %s: %v", i+1, key, err)
			}
		}
		if inbox := testauth.Inbox(t, mailpit, address, 2); len(inbox) != 1 {
			t.Fatalf("%d messages for two requests carrying %s, want 1", len(inbox), key)
		}
		t.Logf("two requests carrying %s → 1 message for %s", key, address)
	})
}

// freshAddress is the per-run address of #61 line 166: a hyphen rather than plus-addressing, and its
// own address per test, because a fixed one collides with the five-per-hour limit and the second
// local run inside an hour would fail on its first request.
func freshAddress(t *testing.T) string {
	t.Helper()
	return "mmr-test-" + uuid.NewV4().String() + "@example.test"
}

func memberRows(t *testing.T, db *pgxpool.Pool, address string) int {
	t.Helper()
	var rows int
	if err := db.QueryRow(t.Context(), `SELECT count(*) FROM member WHERE email = $1`, address).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

// forgetMember drops the row this address made when the test ends, on a context of its own because
// t.Context is already cancelled by then.
func forgetMember(t *testing.T, db *pgxpool.Pool, address string) {
	t.Helper()
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE email = $1`, address); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
}

// redeem is the direct-grant exchange the app makes with a link it was handed (A10): the login token
// on the ios client under the login-token scope. A token already spent answers "" and the refusal.
func redeem(t *testing.T, kc, loginToken string) (access, refusal string) {
	t.Helper()
	return testauth.Exchange(t, kc, url.Values{"grant_type": {"password"}, "client_id": {"ios"}, "scope": {"login-token"}, "login_token": {loginToken}})
}

// loginTokenFromMail is the lt:<uuid> inside the one message Mailpit holds for address. The link is
// {base}?lt={token} and the token is percent-encoded in the body, which Query decodes.
func loginTokenFromMail(t *testing.T, mailpit, address string) string {
	t.Helper()
	inbox := testauth.Inbox(t, mailpit, address, 1)
	if len(inbox) != 1 {
		t.Fatalf("%d messages for %s, want 1", len(inbox), address)
	}
	text := testauth.MessageText(t, mailpit, inbox[0])
	for _, field := range strings.Fields(text) {
		link, err := url.Parse(field)
		if err != nil || !strings.HasPrefix(link.Query().Get("lt"), "lt:") {
			continue
		}
		return link.Query().Get("lt")
	}
	t.Fatalf("no magic link in %q", text)
	return ""
}

// throwaway signs a fresh address in by its magic link, as the app does, and answers its access token: a Member of the
// test's own whose row is dropped at the end, so a test that stamps or deletes never touches the realm's test Member.
func throwaway(t *testing.T, db *pgxpool.Pool, base, kc, mailpit string) string {
	t.Helper()
	address := freshAddress(t)
	forgetMember(t, db, address)
	anonymous := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(""))
	if _, err := anonymous.RequestMagicLink(t.Context(), request(address, "")); err != nil {
		t.Fatal(err)
	}
	access, refused := redeem(t, kc, loginTokenFromMail(t, mailpit, address))
	if access == "" {
		t.Fatalf("redeeming a fresh address's link: %s", refused)
	}
	return access
}
