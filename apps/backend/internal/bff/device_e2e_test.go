package bff

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestRegisterDeviceEndToEnd (#178): the API refuses a keyless registration, then stores the phone's token in notify
// under the signed-in Member, once for a repeated key.
func TestRegisterDeviceEndToEnd(t *testing.T) {
	base, kc, pw, dsn, notifyDSN := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL"), os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" || notifyDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD, MMR_TEST_MEMBERS_DATABASE_URL and NOTIFY_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, notifyDB := pool(ctx, t, dsn), pool(ctx, t, notifyDSN)
	tok := testauth.UserToken(t, kc, pw)
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	token := hex.EncodeToString([]byte(rand.Text()))
	t.Cleanup(func() {
		if _, err := notifyDB.Exec(context.Background(), `DELETE FROM device WHERE apns_token = $1`, token); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	req := connect.NewRequest(&moneyrecapv1.BffServiceRegisterDeviceRequest{ApnsToken: token})
	if _, err := client.RegisterDevice(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("no Idempotency-Key: %v, want invalid_argument", err)
	}
	req.Header().Set("Idempotency-Key", rand.Text())
	for range 2 {
		if _, err := client.RegisterDevice(ctx, req); err != nil {
			t.Fatal(err)
		}
	}
	id := memberOf(ctx, t, db, testauth.Decode(t, tok).Sub)
	var owner string
	var rows int
	if err := notifyDB.QueryRow(ctx, `SELECT min(member_id::text), count(*) FROM device WHERE apns_token = $1`, token).Scan(&owner, &rows); err != nil || rows != 1 || owner != id {
		t.Fatalf("notify holds %d rows for the token, owned by %q (%v), want one row owned by the signed-in Member", rows, owner, err)
	}
}
