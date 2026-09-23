package bff

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestReplayEndToEnd retries SetTimezone through the stack's BFF after deleting the Member's row, which
// members would create again, so no row proves Redis answered; a key reused for another zone is refused (#127).
func TestReplayEndToEnd(t *testing.T) {
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
	forget := func() {
		t.Helper()
		if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, sub); err != nil {
			t.Fatal(err)
		}
	}
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	key := rand.Text()
	send := func(zone string) (*moneyrecapv1.Profile, error) {
		req := connect.NewRequest(&moneyrecapv1.BffServiceSetTimezoneRequest{Timezone: zone})
		req.Header().Set("Idempotency-Key", key)
		res, err := client.SetTimezone(ctx, req)
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}

	forget()
	first, err := send("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := send("UTC"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("the key reused for another zone: %v, want members' invalid_argument", err)
	}
	forget()
	again, err := send("Asia/Jakarta")
	if err != nil || !proto.Equal(first, again) {
		t.Fatalf("the retry: %v, %v; want the first answer %v", again, err, first)
	}
	var rows int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE keycloak_subject = $1`, sub).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("%d Member rows after the retry (%v), want 0: the retry reached members", rows, err)
	}
}
