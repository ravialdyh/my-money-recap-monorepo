package bff

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestEndToEnd drives the BFF's HTTP boundary against the real Compose stack (#17, spec #10).
// It skips unless BFF_TEST_URL, BFF_TEST_KEYCLOAK, BFF_TEST_USER_PASSWORD and
// BFF_TEST_DATABASE_URL are all set; infra/compose/README.md lists their values.
func TestEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("BFF_TEST_KEYCLOAK"), os.Getenv("BFF_TEST_USER_PASSWORD"), os.Getenv("BFF_TEST_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" {
		t.Skip("BFF_TEST_URL, BFF_TEST_KEYCLOAK, BFF_TEST_USER_PASSWORD and BFF_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)

	accessToken := testauth.Token(t, kc, url.Values{"grant_type": {"password"}, "client_id": {"ios"}, "username": {"test"}, "password": {pw}})
	claims := testauth.Decode(t, accessToken)
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, claims.Sub); err != nil {
		t.Fatal(err)
	}

	getHome := func(tok string) (*moneyrecapv1.HomeView, error) {
		res, err := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok)).GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Home, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	mintForeignlySignedToken := func(exp time.Time) string {
		return testauth.Mint(t, key, "not-a-realm-key", jwt.Claims{Issuer: claims.Iss, Subject: claims.Sub, Audience: jwt.Audience{"bff"}, Expiry: jwt.NewNumericDate(exp), IssuedAt: jwt.NewNumericDate(exp.Add(-time.Hour))})
	}
	for name, tok := range map[string]string{
		"no token":                    "",
		"token signed by another key": mintForeignlySignedToken(time.Now().Add(time.Hour)),
		"expired token":               mintForeignlySignedToken(time.Now().Add(-time.Hour)),
	} {
		_, err := getHome(tok)
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("%s: %v, want unauthenticated", name, err)
		}
		t.Logf("%s → %v", name, err)
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("%d Member rows after the unauthenticated calls, want 0: a bad token reached members", rows)
	}

	home, err := getHome(accessToken)
	if err != nil {
		t.Fatal(err)
	}
	assertZeroHome(t, home)
	now := time.Now().UTC()
	if home.Month != now.Format(monthLayout) || home.FirstRecapOn != time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format(dateLayout) {
		t.Fatalf("HomeView dates: %v", home)
	}
	t.Logf("HomeView: %v", home)
	var id, email string
	if err := db.QueryRow(ctx, `SELECT id, email FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&id, &email); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 || email != claims.Email {
		t.Fatalf("%d rows, email %q; want 1 row with %q", rows, email, claims.Email)
	}
	t.Logf("one Member row: id %s, email %q", id, email)

	if _, err := getHome(accessToken); err != nil {
		t.Fatal(err)
	}
	var id2 string
	if err := db.QueryRow(ctx, `SELECT id FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&id2); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 1 || id2 != id {
		t.Fatalf("after the second call: %d rows, id %s (first %s)", rows, id2, id)
	}
	t.Logf("second GetHome: still 1 row, same member id %s", id2)
}
