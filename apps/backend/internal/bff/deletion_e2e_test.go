package bff

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/members"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestLeavingEndToEnd (#183), on a throwaway Member: ForgetDevice deletes the phone's row; a deletion request needs
// its key, answers the stamp and keeps it on a repeat; the event sets notify's flag and Keep my account clears it.
func TestLeavingEndToEnd(t *testing.T) {
	base, kc, mailpit := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_MAILPIT")
	dsn, notifyDSN := os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL"), os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if base == "" || kc == "" || mailpit == "" || dsn == "" || notifyDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_MAILPIT, MMR_TEST_MEMBERS_DATABASE_URL and NOTIFY_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, notifyDB := pool(ctx, t, dsn), pool(ctx, t, notifyDSN)
	access := throwaway(t, db, base, kc, mailpit)
	onboard(ctx, t, base, access)
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(access))
	id := memberOf(ctx, t, db, testauth.Decode(t, access).Sub)
	t.Cleanup(func() {
		for _, table := range []string{"device", "idempotency", "member_facts"} {
			if _, err := notifyDB.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = $1`, id); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
	pending := func() bool {
		var flag bool
		if err := notifyDB.QueryRow(ctx, `SELECT coalesce(bool_or(deletion_pending), false) FROM member_facts WHERE member_id = $1`, id).Scan(&flag); err != nil {
			t.Fatal(err)
		}
		return flag
	}

	token := hex.EncodeToString([]byte(rand.Text()))
	register := connect.NewRequest(&moneyrecapv1.BffServiceRegisterDeviceRequest{ApnsToken: token})
	register.Header().Set("Idempotency-Key", rand.Text())
	if _, err := client.RegisterDevice(ctx, register); err != nil {
		t.Fatal(err)
	}
	forget := connect.NewRequest(&moneyrecapv1.BffServiceForgetDeviceRequest{ApnsToken: token})
	if _, err := client.ForgetDevice(ctx, forget); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("ForgetDevice with no Idempotency-Key: %v, want invalid_argument", err)
	}
	forget.Header().Set("Idempotency-Key", rand.Text())
	if _, err := client.ForgetDevice(ctx, forget); err != nil {
		t.Fatal(err)
	}
	var devices int
	if err := notifyDB.QueryRow(ctx, `SELECT count(*) FROM device WHERE apns_token = $1`, token).Scan(&devices); err != nil || devices != 0 {
		t.Fatalf("notify holds %d rows for the token after Log out forgot it (%v), want none", devices, err)
	}

	ask := connect.NewRequest(&moneyrecapv1.BffServiceRequestDeletionRequest{})
	if _, err := client.RequestDeletion(ctx, ask); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("no Idempotency-Key: %v, want invalid_argument", err)
	}
	ask.Header().Set("Idempotency-Key", rand.Text())
	first, err := client.RequestDeletion(ctx, ask)
	if err != nil || first.Msg.DeletionRequestedAt == nil || time.Since(first.Msg.DeletionRequestedAt.AsTime()) > time.Minute {
		t.Fatalf("RequestDeletion: %v, %v; want the stamp just written", first, err)
	}
	ask.Header().Set("Idempotency-Key", rand.Text())
	again, err := client.RequestDeletion(ctx, ask)
	if err != nil || !again.Msg.DeletionRequestedAt.AsTime().Equal(first.Msg.DeletionRequestedAt.AsTime()) {
		t.Fatalf("a second request: %v, %v; want the first stamp kept", again, err)
	}
	within(t, "members.deletion.requested sets notify's flag", pending)

	keep := connect.NewRequest(&moneyrecapv1.BffServiceCancelDeletionRequest{})
	keep.Header().Set("Idempotency-Key", rand.Text())
	if _, err := client.CancelDeletion(ctx, keep); err != nil {
		t.Fatal(err)
	}
	within(t, "members.deletion.cancelled clears notify's flag", func() bool { return !pending() })
}

// TestDeletionJobEndToEnd (#184), on a throwaway Member with a Spend and a phone, its request stamped 25 hours back:
// members' job runs until ledger and notify hold nothing and both purges are stamped, then once more after
// tokenLifetime to hard-delete.
func TestDeletionJobEndToEnd(t *testing.T) {
	base, kc, mailpit, secret := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_MAILPIT"), os.Getenv("MEMBERS_TEST_CLIENT_SECRET")
	dsn, ledgerDSN, notifyDSN := os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL"), os.Getenv("LEDGER_TEST_DATABASE_URL"), os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if base == "" || kc == "" || mailpit == "" || secret == "" || dsn == "" || ledgerDSN == "" || notifyDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_MAILPIT, MEMBERS_TEST_CLIENT_SECRET and the three database URLs not all set")
	}
	ctx := context.Background()
	db, ledgerDB, notifyDB := pool(ctx, t, dsn), pool(ctx, t, ledgerDSN), pool(ctx, t, notifyDSN)
	access := throwaway(t, db, base, kc, mailpit)
	onboard(ctx, t, base, access)
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(access))
	id := memberOf(ctx, t, db, testauth.Decode(t, access).Sub)
	spend := connect.NewRequest(&moneyrecapv1.BffServiceLogSpendRequest{
		Amount: &moneyrecapv1.Money{Amount: "25000", CurrencyCode: "IDR"}, Category: moneyrecapv1.Category_CATEGORY_SNACKS, Date: time.Now().UTC().Format(dateLayout),
	})
	device := connect.NewRequest(&moneyrecapv1.BffServiceRegisterDeviceRequest{ApnsToken: hex.EncodeToString([]byte(rand.Text()))})
	ask := connect.NewRequest(&moneyrecapv1.BffServiceRequestDeletionRequest{})
	for _, req := range []interface{ Header() http.Header }{spend, device, ask} {
		req.Header().Set("Idempotency-Key", rand.Text())
	}
	if _, err := client.LogSpend(ctx, spend); err != nil {
		t.Fatal(err)
	}
	if _, err := client.RegisterDevice(ctx, device); err != nil {
		t.Fatal(err)
	}
	if _, err := client.RequestDeletion(ctx, ask); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE member SET deletion_requested_at = now() - interval '25 hours' WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	job := &members.Service{DB: db, Keycloak: members.Keycloak{
		Admin: kc + "/admin/realms/mmr", Account: &auth.ServiceAccount{Realm: kc + "/realms/mmr", Client: "members", Secret: secret},
	}}
	if err := job.DeleteDue(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	left := func(db *pgxpool.Pool, tables ...string) int {
		total := 0
		for _, table := range tables {
			var n int
			if err := db.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE member_id = $1`, id).Scan(&n); err != nil {
				t.Fatal(err)
			}
			total += n
		}
		return total
	}
	within(t, "ledger and notify hold nothing and the progress row has both purge stamps", func() bool {
		var stamped int
		if err := db.QueryRow(ctx, `SELECT count(*) FROM deletion_progress WHERE member_id = $1
			AND ledger_purged_at IS NOT NULL AND notify_purged_at IS NOT NULL`, id).Scan(&stamped); err != nil {
			t.Fatal(err)
		}
		return stamped == 1 &&
			left(ledgerDB, "spend", "salary_setting", "saving_target_setting", "currency_conversion", "idempotency", "member_facts") == 0 &&
			left(notifyDB, "notification_preference", "device", "month_flag", "sent_log", "idempotency", "member_facts") == 0
	})
	if err := job.DeleteDue(ctx, time.Now().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var member int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM member WHERE id = $1`, id).Scan(&member); err != nil {
		t.Fatal(err)
	}
	if rows := member + left(db, "deletion_progress", "outbox", "idempotency"); rows != 0 {
		t.Errorf("%d rows about the Member in members once the tokens have expired, want none", rows)
	}
}
