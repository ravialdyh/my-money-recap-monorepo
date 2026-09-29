package notify

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestAZoneChangeReachesNotify changes the test Member's zone in the members container and never calls notify: the
// outbox, the relay, the broker and notify.members carry each change into notify's copy. It leaves Asia/Jakarta.
func TestAZoneChangeReachesNotify(t *testing.T) {
	membersURL, kc, pw := os.Getenv("MEMBERS_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD")
	if membersURL == "" || kc == "" || pw == "" {
		t.Skip("MEMBERS_TEST_URL, MMR_TEST_KEYCLOAK and MMR_TEST_USER_PASSWORD not all set")
	}
	s := &Service{DB: testDB(t)}
	ctx := context.Background()
	tok := testauth.UserToken(t, kc, pw)
	claims := testauth.Decode(t, tok)
	members := moneyrecapv1connect.NewMembersServiceClient(http.DefaultClient, membersURL, testauth.Bearer(tok))
	me, err := members.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: claims.Sub, Email: claims.Email}))
	if err != nil {
		t.Fatal(err)
	}
	id := me.Msg.Profile.MemberId
	for _, zone := range []string{"Pacific/Auckland", "Asia/Jakarta"} {
		if _, err := members.SetTimezone(ctx, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: rand.Text(), MemberId: id, Timezone: zone})); err != nil {
			t.Fatal(err)
		}
		var held string
		for deadline := time.Now().Add(15 * time.Second); held != zone; time.Sleep(200 * time.Millisecond) {
			if time.Now().After(deadline) {
				t.Fatalf("after 15 s notify's copy holds zone %q, want %q", held, zone)
			}
			if err := s.DB.QueryRow(ctx, `SELECT coalesce(timezone, '') FROM member_facts WHERE member_id = $1`, id).Scan(&held); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				t.Fatal(err)
			}
		}
	}
	t.Log("two zone changes in members reached notify's copy, with no call to notify")
}
