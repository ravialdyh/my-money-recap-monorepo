package members

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// Both tests need the Compose stack and skip without the MEMBERS_TEST_* variables that
// infra/compose/README.md lists. MEMBERS_TEST_BFF_CLIENT_SECRET is optional and adds the
// wrong-audience case.

func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("MEMBERS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("MEMBERS_TEST_DATABASE_URL not set")
	}
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := Migrate(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestConcurrentFirstCallsCreateOneRow(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	subject := "concurrent-" + t.Name() + "-" + strings.ReplaceAll(t.TempDir(), "/", "")
	t.Cleanup(func() { db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, subject) })

	const n = 16
	ids := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var m Row
			m, errs[i] = Upsert(ctx, db, subject, "race@example.com")
			ids[i] = m.ID
		}()
	}
	wg.Wait()
	for i := range n {
		if errs[i] != nil {
			t.Fatalf("call %d: %v", i, errs[i])
		}
		if ids[i] != ids[0] {
			t.Fatalf("call %d returned id %s, call 0 returned %s", i, ids[i], ids[0])
		}
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, subject).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 {
		t.Fatalf("%d rows for one subject, want 1", rows)
	}
	t.Logf("%d concurrent first calls → 1 row, every call returned member id %s", n, ids[0])
}

func TestConnectBoundary(t *testing.T) {
	base := os.Getenv("MEMBERS_TEST_URL")
	if base == "" || os.Getenv("MEMBERS_TEST_KEYCLOAK") == "" || os.Getenv("MEMBERS_TEST_USER_PASSWORD") == "" {
		t.Skip("MEMBERS_TEST_URL, MEMBERS_TEST_KEYCLOAK and MEMBERS_TEST_USER_PASSWORD not all set")
	}
	db := testDB(t)
	ctx := context.Background()
	kc := os.Getenv("MEMBERS_TEST_KEYCLOAK")
	tok := testauth.Token(t, kc, url.Values{"grant_type": {"password"}, "client_id": {"ios"}, "username": {"test"}, "password": {os.Getenv("MEMBERS_TEST_USER_PASSWORD")}})
	claims := testauth.Decode(t, tok)
	sub, email := claims.Sub, claims.Email
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, sub); err != nil {
		t.Fatal(err)
	}

	client := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base)
	authed := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	getOrCreate := func(c moneyrecapv1connect.MembersServiceClient, subject, email string) (*moneyrecapv1.Profile, error) {
		res, err := c.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: subject, Email: email}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}

	first, err := getOrCreate(authed, sub, email)
	if err != nil {
		t.Fatal(err)
	}
	if first.MemberId == "" || first.Email != email {
		t.Fatalf("first call: %v", first)
	}
	var rows int
	db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, sub).Scan(&rows)
	if rows != 1 {
		t.Fatalf("%d rows after first call, want 1", rows)
	}
	second, err := getOrCreate(authed, sub, email)
	if err != nil {
		t.Fatal(err)
	}
	if second.MemberId != first.MemberId {
		t.Fatalf("second call returned %s, first %s", second.MemberId, first.MemberId)
	}
	t.Logf("first and second call: member id %s", first.MemberId)

	third, err := getOrCreate(authed, sub, "renamed@example.com")
	if err != nil {
		t.Fatal(err)
	}
	var stored string
	db.QueryRow(ctx, `SELECT email FROM member WHERE id = $1`, first.MemberId).Scan(&stored)
	if third.MemberId != first.MemberId || third.Email != "renamed@example.com" || stored != "renamed@example.com" {
		t.Fatalf("email refresh: response %v, stored %q", third, stored)
	}
	t.Logf("different email: stored email is now %q", stored)

	prof, err := authed.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: first.MemberId}))
	if err != nil {
		t.Fatal(err)
	}
	if p := prof.Msg.Profile; p.MemberId != first.MemberId || p.Email != "renamed@example.com" || p.DisplayName != "" || p.Timezone != "" {
		t.Fatalf("GetProfile: %v", p)
	}
	t.Logf("GetProfile: %v", prof.Msg.Profile)

	if _, err := getOrCreate(authed, "someone-else", email); connect.CodeOf(err) != connect.CodePermissionDenied {
		t.Errorf("subject mismatch: %v, want permission_denied", err)
	}
	unauth := map[string]moneyrecapv1connect.MembersServiceClient{
		"no token":      client,
		"garbage token": moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base, testauth.Bearer("not.a.jwt")),
	}
	if secret := os.Getenv("MEMBERS_TEST_BFF_CLIENT_SECRET"); secret != "" {
		bffTok := testauth.Token(t, kc, url.Values{"grant_type": {"client_credentials"}, "client_id": {"bff"}, "client_secret": {secret}})
		unauth["wrong audience"] = moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, base, testauth.Bearer(bffTok))
	}
	for name, c := range unauth {
		if _, err := getOrCreate(c, sub, email); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("GetOrCreateBySubject %s: %v, want unauthenticated", name, err)
		}
		if _, err := c.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: first.MemberId})); connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("GetProfile %s: %v, want unauthenticated", name, err)
		}
		t.Logf("%s: both RPCs unauthenticated", name)
	}
}
