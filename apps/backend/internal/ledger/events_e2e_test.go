package ledger

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestMembersEventsReachTheReplica writes to the members container and never calls ledger: the outbox, the
// relay, the broker and ledger's consumer carry every staged event into processed_events and the replica.
func TestMembersEventsReachTheReplica(t *testing.T) {
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
	forget(t, db, id)
	zoned, err := members.SetTimezone(ctx, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: rand.Text(), MemberId: id, Timezone: "Asia/Jakarta"}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := members.SaveProfile(ctx, connect.NewRequest(&moneyrecapv1.SaveProfileRequest{
		IdempotencyKey: rand.Text(), MemberId: id, CurrencyCode: proto.String("IDR"), OnboardingStep: proto.Int32(2),
	})); err != nil {
		t.Fatal(err)
	}
	rows, err := membersDB.Query(ctx, `SELECT id::text FROM outbox WHERE member_id = $1`, id)
	if err != nil {
		t.Fatal(err)
	}
	staged, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil || len(staged) != 3 {
		t.Fatalf("members staged %d events (%v) for a creation, a zone and a Currency, want 3", len(staged), err)
	}
	want := "- Asia/Jakarta " + zoned.Msg.Profile.SignedUpOn + " IDR"
	var processed int
	for deadline := time.Now().Add(15 * time.Second); processed != 3 || replica(t, db, id) != want; time.Sleep(200 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatalf("after 15 s ledger processed %d of the 3 events and holds %q, want %q", processed, replica(t, db, id), want)
		}
		if err := db.QueryRow(ctx, `SELECT count(*) FROM processed_events WHERE message_id = ANY($1)`, staged).Scan(&processed); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("3 events from members reached ledger's replica, with no call to ledger: %q", want)
}
