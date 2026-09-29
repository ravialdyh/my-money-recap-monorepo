package bff

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestNotificationPrefsEndToEnd (#179): a Member of its own, signed in by magic link so the shared test Member's
// 120-a-minute limit is untouched, reads PRD §10's defaults; a keyless save is refused, a keyed one stored once.
func TestNotificationPrefsEndToEnd(t *testing.T) {
	base, kc, mailpit, dsn, notifyDSN := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_MAILPIT"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL"), os.Getenv("NOTIFY_TEST_DATABASE_URL")
	if base == "" || kc == "" || mailpit == "" || dsn == "" || notifyDSN == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_MAILPIT, MMR_TEST_MEMBERS_DATABASE_URL and NOTIFY_TEST_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, notifyDB := pool(ctx, t, dsn), pool(ctx, t, notifyDSN)
	address := freshAddress(t)
	forgetMember(t, db, address)
	anonymous := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(""))
	if _, err := anonymous.RequestMagicLink(ctx, request(address, "")); err != nil {
		t.Fatal(err)
	}
	tok, refused := redeem(t, kc, loginTokenFromMail(t, mailpit, address))
	if tok == "" {
		t.Fatalf("redeeming the link: %s", refused)
	}
	onboard(ctx, t, base, tok)
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	read := func() *moneyrecapv1.NotificationPrefs {
		t.Helper()
		res, err := client.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetSettingsRequest{}))
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg.Settings.NotificationPrefs
	}
	id := memberOf(ctx, t, db, testauth.Decode(t, tok).Sub)
	t.Cleanup(func() {
		if _, err := notifyDB.Exec(context.Background(), `DELETE FROM notification_preference WHERE member_id = $1`, id); err != nil {
			t.Logf("clearing the preferences: %v", err)
		}
	})

	defaults := &moneyrecapv1.NotificationPrefs{RecapReady: true, ReminderTime: "20:00", TargetWarning: true}
	if got := read(); !proto.Equal(got, defaults) {
		t.Fatalf("before any save Settings reads %v, want %v", got, defaults)
	}
	reminder := &moneyrecapv1.NotificationPrefs{RecapReady: true, DailyReminder: true, ReminderTime: "21:15", TargetWarning: true}
	req := connect.NewRequest(&moneyrecapv1.BffServiceSaveNotificationPrefsRequest{Prefs: reminder})
	if _, err := client.SaveNotificationPrefs(ctx, req); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("no Idempotency-Key: %v, want invalid_argument", err)
	}
	req.Header().Set("Idempotency-Key", rand.Text())
	for range 2 {
		res, err := client.SaveNotificationPrefs(ctx, req)
		if err != nil || !proto.Equal(res.Msg.Prefs, reminder) {
			t.Fatalf("saving the reminder: %v (%v), want it answered as sent", res, err)
		}
	}
	if got := read(); !proto.Equal(got, reminder) {
		t.Errorf("after the save Settings reads %v, want %v", got, reminder)
	}
	var rows int
	if err := notifyDB.QueryRow(ctx, `SELECT count(*) FROM notification_preference WHERE member_id = $1`, id).Scan(&rows); err != nil || rows != 1 {
		t.Errorf("notify holds %d rows for the Member (%v), want one", rows, err)
	}
}
