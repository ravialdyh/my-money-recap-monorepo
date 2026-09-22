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

// TestSetTimezoneEndToEnd sends Asia/Jakarta through the BFF to the members container, whose image has
// no zone files, so the answer proves the binary carries Go's zone data (#96); a keyless call creates
// nothing. It skips like TestEndToEnd.
func TestSetTimezoneEndToEnd(t *testing.T) {
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
	sub := testauth.Decode(t, tok).Sub
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, sub); err != nil {
		t.Fatal(err)
	}
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	send := func(key, zone string) (*moneyrecapv1.Profile, error) {
		req := connect.NewRequest(&moneyrecapv1.BffServiceSetTimezoneRequest{Timezone: zone})
		if key != "" {
			req.Header().Set("Idempotency-Key", key)
		}
		res, err := client.SetTimezone(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}

	if _, err := send("", "Asia/Jakarta"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("no key: %v, want invalid_argument", err)
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, sub).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("%d Member rows after a keyless call (%v), want 0", rows, err)
	}
	p, err := send(rand.Text(), "Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	var created time.Time
	if err := db.QueryRow(ctx, `SELECT created_at FROM member WHERE keycloak_subject = $1`, sub).Scan(&created); err != nil {
		t.Fatal(err)
	}
	jakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	if p.Timezone != "Asia/Jakarta" || p.SignedUpOn != created.In(jakarta).Format(time.DateOnly) {
		t.Fatalf("SetTimezone: %v, want Asia/Jakarta and the Jakarta date of %v", p, created)
	}
	t.Logf("the members container accepted Asia/Jakarta: %v", p)
	if _, err := send(rand.Text(), "Mars/Olympus_Mons"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("unknown zone: %v, want invalid_argument", err)
	}
}
