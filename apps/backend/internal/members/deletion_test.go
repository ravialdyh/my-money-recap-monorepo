package members

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testdb"
)

// The deletion job (§7.1 steps 3 to 6) on a throwaway Keycloak account: a 25-hour stamp is tombstoned once, its
// account deleted, its subject refused, and both purged events, repeated, hard-delete it; a 23-hour one stays.
func TestTheDeletionJob(t *testing.T) {
	kc, secret := os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MEMBERS_TEST_CLIENT_SECRET")
	if kc == "" || secret == "" {
		t.Skip("MMR_TEST_KEYCLOAK and MEMBERS_TEST_CLIENT_SECRET not both set")
	}
	db := testdb.Scratch(t, "MMR_TEST_MEMBERS_DATABASE_URL", Migrate)
	k := Keycloak{Admin: kc + "/admin/realms/mmr", Account: &auth.ServiceAccount{Realm: kc + "/realms/mmr", Client: "members", Secret: secret}}
	ctx, now := context.Background(), time.Now()
	subject := throwaway(t, k)
	leaving, err := Upsert(ctx, db, subject, "leaving@example.test")
	if err != nil {
		t.Fatal(err)
	}
	staying, err := Upsert(ctx, db, "staying-"+t.Name(), "staying@example.test")
	if err != nil {
		t.Fatal(err)
	}
	for id, ago := range map[string]time.Duration{leaving.ID: 25 * time.Hour, staying.ID: 23 * time.Hour} {
		if _, err := db.Exec(ctx, `UPDATE member SET deletion_requested_at = $2 WHERE id = $1`, id, now.Add(-ago)); err != nil {
			t.Fatal(err)
		}
	}
	s := &Service{DB: db, Keycloak: k}
	for range 2 {
		if err := s.DeleteDue(ctx, now); err != nil {
			t.Fatal(err)
		}
	}
	var tombstoned, untouched, held bool
	if err := db.QueryRow(ctx, `SELECT
		(SELECT deleted_at IS NOT NULL FROM member WHERE id = $1),
		(SELECT deleted_at IS NULL FROM member WHERE id = $2) AND NOT EXISTS (SELECT 1 FROM deletion_progress WHERE member_id = $2),
		(SELECT keycloak_deleted_at IS NULL FROM deletion_progress WHERE member_id = $1)`, leaving.ID, staying.ID).Scan(&tombstoned, &untouched, &held); err != nil {
		t.Fatal(err)
	}
	if !tombstoned || !untouched || held {
		t.Fatalf("after two runs: tombstoned %v, the younger stamp untouched %v, Keycloak still held %v", tombstoned, untouched, held)
	}
	deleted := 0
	for _, e := range outbox(t, db, leaving.ID) {
		if e.key == "members.deleted" {
			deleted++
		}
	}
	if deleted != 1 {
		t.Errorf("%d members.deleted staged over two runs, want 1", deleted)
	}
	if status := admin(t, k, http.MethodGet, "/users/"+subject, ""); status != http.StatusNotFound {
		t.Errorf("the Keycloak account after the job answered %d, want 404", status)
	}
	if _, err := Upsert(ctx, db, subject, ""); !errors.Is(err, errTombstoned) {
		t.Errorf("signing in under the tombstoned subject: %v, want refused", err)
	}

	purge := func(key string) {
		t.Helper()
		body, err := proto.Marshal(&moneyrecapv1.LedgerMemberPurged{Envelope: &moneyrecapv1.EventEnvelope{MemberId: leaving.ID}})
		if err != nil {
			t.Fatal(err)
		}
		if err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
			return ApplyPurged(ctx, tx, amqp.Delivery{RoutingKey: key, Body: body})
		}); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
	}
	rows := func() int {
		t.Helper()
		var n int
		if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM member WHERE id = $1) + (SELECT count(*) FROM deletion_progress WHERE member_id = $1)
			+ (SELECT count(*) FROM outbox WHERE member_id = $1) + (SELECT count(*) FROM idempotency WHERE member_id = $1)`, leaving.ID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	purge("ledger.member.purged")
	purge("ledger.member.purged")
	if rows() == 0 {
		t.Fatal("hard-deleted before notify's purge")
	}
	purge("notify.member.purged")
	if n := rows(); n != 0 {
		t.Errorf("%d rows left about the Member once all three stamps are present, want none", n)
	}
	purge("notify.member.purged")
}

// A deletion still incomplete an hour after its tombstone writes a log line on every run, standing in for A11's alert.
func TestAStalledDeletionIsLogged(t *testing.T) {
	db := testdb.Scratch(t, "MMR_TEST_MEMBERS_DATABASE_URL", Migrate)
	ctx, now := context.Background(), time.Now()
	stalled, fresh := "00000000-0000-4000-8000-00000000e001", "00000000-0000-4000-8000-00000000e002"
	if _, err := db.Exec(ctx, `INSERT INTO deletion_progress (member_id, keycloak_subject, deleted_at, keycloak_deleted_at)
		VALUES ($1, 'stalled', $3, $3), ($2, 'fresh', $4, $4)`, stalled, fresh, now.Add(-2*time.Hour), now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	var logged bytes.Buffer
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	if err := (&Service{DB: db}).DeleteDue(ctx, now); err != nil {
		t.Fatal(err)
	}
	if line := logged.String(); !strings.Contains(line, stalled+" is incomplete an hour after") || strings.Contains(line, fresh) {
		t.Errorf("the log %q, want a line for the stalled deletion alone", line)
	}
}
