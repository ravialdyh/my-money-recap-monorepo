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

// TestCancelDeletionEndToEnd writes a deletion stamp straight into the row: the launch call answers it,
// "Keep my account" clears it through members, and a stamp past the 24 hours stays (D60, #105).
func TestCancelDeletionEndToEnd(t *testing.T) {
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
	keyed := func(req interface{ Header() http.Header }) {
		req.Header().Set("Idempotency-Key", rand.Text())
	}
	launch := func() *moneyrecapv1.Profile {
		t.Helper()
		req := connect.NewRequest(&moneyrecapv1.BffServiceSetTimezoneRequest{Timezone: "Asia/Jakarta"})
		keyed(req)
		res, err := client.SetTimezone(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg.Profile
	}
	keep := func() (*moneyrecapv1.Profile, error) {
		req := connect.NewRequest(&moneyrecapv1.BffServiceCancelDeletionRequest{})
		keyed(req)
		res, err := client.CancelDeletion(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}
	stamp := func(ago string) {
		t.Helper()
		if _, err := db.Exec(ctx, `UPDATE member SET deletion_requested_at = now() - $2::interval WHERE keycloak_subject = $1`, sub, ago); err != nil {
			t.Fatal(err)
		}
	}

	launch()
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `UPDATE member SET deletion_requested_at = NULL WHERE keycloak_subject = $1`, sub); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	stamp("1 hour")
	if p := launch(); p.DeletionRequestedAt == nil || time.Since(p.DeletionRequestedAt.AsTime()) > 2*time.Hour {
		t.Fatalf("the launch call after asking an hour ago: %v, want the stamp", p.DeletionRequestedAt)
	}
	if _, err := client.CancelDeletion(ctx, connect.NewRequest(&moneyrecapv1.BffServiceCancelDeletionRequest{})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no key: %v, want invalid_argument", err)
	}
	if p, err := keep(); err != nil || p.DeletionRequestedAt != nil {
		t.Fatalf("Keep my account: %v, %v; want the Profile with no stamp", p, err)
	}
	var cleared bool
	if err := db.QueryRow(ctx, `SELECT deletion_requested_at IS NULL FROM member WHERE keycloak_subject = $1`, sub).Scan(&cleared); err != nil || !cleared {
		t.Fatalf("the stored stamp after Keep my account: cleared %v, %v", cleared, err)
	}
	if p := launch(); p.DeletionRequestedAt != nil {
		t.Errorf("the next launch still carries a stamp: %v", p.DeletionRequestedAt)
	}
	stamp("25 hours")
	if _, err := keep(); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("a stamp 25 hours old: %v, want failed_precondition", err)
	}
}
