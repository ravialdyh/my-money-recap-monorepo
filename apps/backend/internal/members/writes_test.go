package members

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// A Member created at 20:00 UTC on 22 September signed up on the 23rd in Jakarta. A repeated key replays
// the first answer and writes nothing, a reused key with another zone is refused, and a later zone
// changes only the zone (A3, A14).
func TestSetTimezoneStampsTheLocalSignUpDateOnceAndHonoursTheKey(t *testing.T) {
	db := testDB(t)
	subject := "timezone-" + t.Name()
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject})
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	m, err := Upsert(ctx, db, subject, "zone@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE member SET created_at = '2026-09-22T20:00:00Z' WHERE id = $1`, m.ID); err != nil {
		t.Fatal(err)
	}
	s := &Service{DB: db}
	set := func(key, zone string) (*moneyrecapv1.Profile, error) {
		res, err := s.SetTimezone(ctx, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: key, MemberId: m.ID, Timezone: zone}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}
	stored := func() (zone, signedUp string, keys int) {
		t.Helper()
		if err := db.QueryRow(ctx, `SELECT timezone, to_char(signed_up_on, 'YYYY-MM-DD'), (SELECT count(*) FROM idempotency WHERE member_id = $1) FROM member WHERE id = $1`, m.ID).Scan(&zone, &signedUp, &keys); err != nil {
			t.Fatal(err)
		}
		return zone, signedUp, keys
	}

	first, err := set("key-1", "Asia/Jakarta")
	if err != nil || first.Timezone != "Asia/Jakarta" || first.SignedUpOn != "2026-09-23" {
		t.Fatalf("first SetTimezone: %v, %v, want Asia/Jakarta signed up 2026-09-23, the Jakarta date of 20:00 UTC", first, err)
	}
	if _, err := db.Exec(ctx, `UPDATE member SET timezone = 'Europe/Paris' WHERE id = $1`, m.ID); err != nil {
		t.Fatal(err)
	}
	again, err := set("key-1", "Asia/Jakarta")
	if zone, _, keys := stored(); err != nil || again.Timezone != "Asia/Jakarta" || zone != "Europe/Paris" || keys != 1 {
		t.Fatalf("the same key again: %v, %v, stored zone %q, %d keys; want the first answer, Europe/Paris untouched, 1 key", again, err, zone, keys)
	}
	if _, err := set("key-1", "Europe/London"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("key-1 reused for another zone: %v, want invalid_argument", err)
	}
	later, err := set("key-2", "Europe/London")
	if zone, signedUp, keys := stored(); err != nil || later.Timezone != "Europe/London" || zone != "Europe/London" || signedUp != "2026-09-23" || keys != 2 {
		t.Fatalf("a later zone: %v, %v, stored %q signed up %q, %d keys; want Europe/London, the date kept, 2 keys", later, err, zone, signedUp, keys)
	}

	for name, req := range map[string]*moneyrecapv1.SetTimezoneRequest{
		"an unknown zone": {IdempotencyKey: "key-3", MemberId: m.ID, Timezone: "Mars/Olympus_Mons"},
		"no zone":         {IdempotencyKey: "key-3", MemberId: m.ID},
		"Local":           {IdempotencyKey: "key-3", MemberId: m.ID, Timezone: "Local"},
		"no key":          {MemberId: m.ID, Timezone: "Asia/Jakarta"},
	} {
		if _, err := s.SetTimezone(ctx, connect.NewRequest(req)); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", name, err)
		}
	}
	other := auth.NewContext(context.Background(), auth.Claims{Subject: "someone-else"})
	if _, err := s.SetTimezone(other, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: "key-4", MemberId: m.ID, Timezone: "Asia/Jakarta"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another token's Member: %v, want not_found", err)
	}
	if zone, _, keys := stored(); zone != "Europe/London" || keys != 2 {
		t.Fatalf("after the refused calls: stored %q with %d keys, want Europe/London and 2", zone, keys)
	}
}

// A key older than 24 hours is swept by the next keyed write, so a day-old key no longer replays (A14).
func TestKeysOlderThanADayAreSwept(t *testing.T) {
	db := testDB(t)
	subject := "sweep-" + t.Name()
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject})
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	m, err := Upsert(ctx, db, subject, "sweep@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO idempotency (member_id, idempotency_key, request_hash, created_at) VALUES ($1, 'old', '\x00', now() - interval '25 hours'), ($1, 'recent', '\x00', now() - interval '23 hours')`, m.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := (&Service{DB: db}).SetTimezone(ctx, connect.NewRequest(&moneyrecapv1.SetTimezoneRequest{IdempotencyKey: "new", MemberId: m.ID, Timezone: "Asia/Jakarta"})); err != nil {
		t.Fatal(err)
	}
	var keys []string
	if err := db.QueryRow(ctx, `SELECT array_agg(idempotency_key ORDER BY idempotency_key) FROM idempotency WHERE member_id = $1`, m.ID).Scan(&keys); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != "new" || keys[1] != "recent" {
		t.Fatalf("keys after a write: %v, want [new recent] with the 25-hour-old one swept", keys)
	}
}
