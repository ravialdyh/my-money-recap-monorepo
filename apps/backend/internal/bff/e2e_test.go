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

// TestEndToEnd is the backend end-to-end test (#17, spec #10): the BFF's HTTP boundary against
// the real Compose stack. It skips unless all of these are set:
//   BFF_TEST_URL            http://localhost:8082
//   BFF_TEST_KEYCLOAK       http://localhost:8080
//   BFF_TEST_USER_PASSWORD  the realm's test user password
//   BFF_TEST_DATABASE_URL   postgres://members:<pw>@localhost:5432/members, to count Member rows

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

	// Sign in as the test user through the direct grant, as the phone's test seam does.
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

	// Tokens the realm never issued: right issuer, audience and subject, but signed by another key.
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	mint := func(exp time.Time) string {
		return testauth.Mint(t, key, "not-a-realm-key", jwt.Claims{Issuer: claims.Iss, Subject: claims.Sub, Audience: jwt.Audience{"bff"}, Expiry: jwt.NewNumericDate(exp), IssuedAt: jwt.NewNumericDate(exp.Add(-time.Hour))})
	}
	for name, tok := range map[string]string{
		"no token":                    "",
		"token signed by another key": mint(time.Now().Add(time.Hour)),
		"expired token":               mint(time.Now().Add(-time.Hour)),
	} {
		_, err := getHome(tok)
		if connect.CodeOf(err) != connect.CodeUnauthenticated {
			t.Errorf("%s: %v, want unauthenticated", name, err)
		}
		t.Logf("%s → %v", name, err)
	}
	var rows int
	db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&rows)
	if rows != 0 {
		t.Fatalf("%d Member rows after the unauthenticated calls, want 0: a bad token reached members", rows)
	}

	// The test user's token: a HomeView of zeros and exactly one Member row.
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
	db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&rows)
	if rows != 1 || email != claims.Email {
		t.Fatalf("%d rows, email %q; want 1 row with %q", rows, email, claims.Email)
	}
	t.Logf("one Member row: id %s, email %q", id, email)

	// A second GetHome finds the same Member and creates no second row.
	if _, err := getHome(accessToken); err != nil {
		t.Fatal(err)
	}
	var id2 string
	db.QueryRow(ctx, `SELECT id FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&id2)
	db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, claims.Sub).Scan(&rows)
	if rows != 1 || id2 != id {
		t.Fatalf("after the second call: %d rows, id %s (first %s)", rows, id2, id)
	}
	t.Logf("second GetHome: still 1 row, same member id %s", id2)
}
