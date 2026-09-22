package ledger

import (
	"context"
	"crypto/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestConnectBoundary calls the ledger container with the realm test Member's token: the replica fills from
// members under it, and another audience's token is refused (§4.2 point 3). Skips without its variables.
func TestConnectBoundary(t *testing.T) {
	base, membersURL, kc, pw := os.Getenv("LEDGER_TEST_URL"), os.Getenv("MEMBERS_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD")
	membersDSN := os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	if base == "" || membersURL == "" || kc == "" || pw == "" || membersDSN == "" {
		t.Skip("LEDGER_TEST_URL, MEMBERS_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD and MMR_TEST_MEMBERS_DATABASE_URL not all set")
	}
	db := testDB(t)
	ctx := context.Background()
	membersDB, err := pgxpool.New(ctx, membersDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(membersDB.Close)
	tok := testauth.UserToken(t, kc, pw)
	claims := testauth.Decode(t, tok)
	if _, err := membersDB.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, claims.Sub); err != nil {
		t.Fatal(err)
	}
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, membersURL, testauth.Bearer(tok))
	created, err := members.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: claims.Sub, Email: claims.Email}))
	if err != nil {
		t.Fatal(err)
	}
	id := created.Msg.Profile.MemberId
	if _, err := members.SetTimezone(ctx, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: rand.Text(), MemberId: id, Timezone: "Asia/Jakarta"})); err != nil {
		t.Fatal(err)
	}
	if _, err := membersDB.Exec(ctx, `UPDATE member SET currency_code = 'IDR' WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, table := range []string{"member_facts", "salary_setting", "saving_target_setting", "idempotency"} {
			if _, err := db.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = $1`, id); err != nil {
				t.Logf("cleanup: %v", err)
			}
		}
	})

	ledger := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	res, err := ledger.SaveSalary(ctx, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{
		IdempotencyKey: rand.Text(), MemberId: id, Amount: &moneyrecapv1.Money{Amount: "3000000", CurrencyCode: "IDR"}, Payday: "25",
	}))
	if err != nil {
		t.Fatal(err)
	}
	var subject, zone, code string
	if err := db.QueryRow(ctx, `SELECT keycloak_subject, timezone, currency_code FROM member_facts WHERE member_id = $1`, id).Scan(&subject, &zone, &code); err != nil {
		t.Fatal(err)
	}
	if subject != claims.Sub || zone != "Asia/Jakarta" || code != "IDR" || res.Msg.Settings.GetSalary().GetAmount().GetAmount() != "3000000" {
		t.Fatalf("replica %q %q %q, settings %v; want the token's subject, Asia/Jakarta, IDR and the Salary", subject, zone, code, res.Msg.Settings)
	}
	t.Logf("SaveSalary through the container, replica filled from members: %v", res.Msg.Settings)

	if secret := os.Getenv("MEMBERS_TEST_BFF_CLIENT_SECRET"); secret != "" {
		bffTok := testauth.Token(t, kc, url.Values{"grant_type": {"client_credentials"}, "client_id": {"bff"}, "client_secret": {secret}})
		other := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, base, testauth.Bearer(bffTok))
		if _, err := other.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: id})); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("a token for another audience: %v, want unauthenticated", err)
		}
	}
	anonymous := moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, base)
	if _, err := anonymous.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: id})); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("no token: %v, want unauthenticated", err)
	}
}

// The ledger role owns its own database and reaches no other: Postgres grants CONNECT to PUBLIC, and the
// init script takes it back (ARCHITECTURE A9).
func TestTheLedgerRoleReachesNoOtherDatabase(t *testing.T) {
	dsn := os.Getenv("LEDGER_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("LEDGER_TEST_DATABASE_URL not set")
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	for _, other := range []string{"members", "notify", "keycloak"} {
		u.Path = "/" + other
		conn, err := pgx.Connect(context.Background(), u.String())
		if err == nil {
			_ = conn.Close(context.Background())
			t.Errorf("the ledger role connected to %s", other)
			continue
		}
		if !strings.Contains(err.Error(), "permission denied for database") {
			t.Errorf("connecting to %s: %v, want permission denied", other, err)
		}
	}
}
