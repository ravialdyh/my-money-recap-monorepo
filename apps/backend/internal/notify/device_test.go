package notify

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// A Member's phones each hold a row, a repeated key writes once, and a token another Member signs in with moves to
// them; a keyless, non-hex or someone else's registration is refused and writes nothing (spec #169, "Devices").
func TestADeviceIsUpsertedOnItsToken(t *testing.T) {
	ann, bob := "00000000-0000-4000-8000-00000000d001", "00000000-0000-4000-8000-00000000d002"
	s := warningService(t, ann, map[string]string{"ann": ann, "bob": bob})
	t.Cleanup(func() {
		for _, table := range []string{"device", "idempotency", "member_facts"} {
			if _, err := s.DB.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = ANY($1)`, []string{ann, bob}); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
	register := func(ctx context.Context, key, id, token string) error {
		_, err := s.RegisterDevice(ctx, connect.NewRequest(&moneyrecapv1.RegisterDeviceRequest{IdempotencyKey: key, MemberId: id, ApnsToken: token}))
		return err
	}
	owner := func(token string) string {
		t.Helper()
		var id string
		if err := s.DB.QueryRow(t.Context(), `SELECT coalesce((SELECT member_id::text FROM device WHERE apns_token = $1), '')`, token).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	phone, pad := "d00dfeed01", "d00dfeed02"

	for _, key := range []string{"k1", "k1"} {
		if err := register(as("ann"), key, ann, "D00DFEED01"); err != nil {
			t.Fatalf("ann's phone under %s: %v", key, err)
		}
	}
	if err := register(as("ann"), "k2", ann, pad); err != nil {
		t.Fatal(err)
	}
	var held int
	if err := s.DB.QueryRow(t.Context(), `SELECT count(*) FROM device WHERE member_id = $1`, ann).Scan(&held); err != nil || held != 2 {
		t.Fatalf("ann holds %d devices (%v), want her phone and her iPad", held, err)
	}
	if got := owner(phone); got != ann {
		t.Fatalf("the phone belongs to %q, want ann, its token stored as lowercase hex", got)
	}
	if err := register(as("bob"), "k1", bob, phone); err != nil {
		t.Fatalf("bob signing in on ann's phone: %v", err)
	}
	if got := owner(phone); got != bob {
		t.Errorf("after bob registers it the phone belongs to %q, want bob", got)
	}
	if got := owner(pad); got != ann {
		t.Errorf("ann's iPad belongs to %q, want ann still", got)
	}
	for name, call := range map[string]func() error{
		"no key":         func() error { return register(as("ann"), "", ann, "abc1") },
		"not hex":        func() error { return register(as("ann"), "k9", ann, "not-a-token") },
		"no token":       func() error { return register(as("ann"), "k9", ann, "") },
		"another Member": func() error { return register(as("bob"), "k9", ann, "abc1") },
	} {
		if err := call(); err == nil {
			t.Errorf("%s: registered, want refused", name)
		}
	}
	if got := owner("abc1"); got != "" {
		t.Errorf("a refused registration stored a row for %q", got)
	}
}

// Log out forgets the phone: this Member's row for the token goes, once for a repeated key, while another Member's
// row for a token is left as it is; a keyless or someone else's call is refused (spec #169, "Devices").
func TestForgetDeviceDeletesOnlyTheMembersRow(t *testing.T) {
	ann, bob := "00000000-0000-4000-8000-00000000d011", "00000000-0000-4000-8000-00000000d012"
	s := warningService(t, ann, map[string]string{"ann": ann, "bob": bob})
	t.Cleanup(func() {
		for _, table := range []string{"device", "idempotency", "member_facts"} {
			if _, err := s.DB.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = ANY($1)`, []string{ann, bob}); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
	if _, err := s.DB.Exec(t.Context(), `INSERT INTO device (apns_token, member_id) VALUES ('f0f0', $1), ('b0b0', $2)`, ann, bob); err != nil {
		t.Fatal(err)
	}
	forget := func(ctx context.Context, key, id, token string) error {
		_, err := s.ForgetDevice(ctx, connect.NewRequest(&moneyrecapv1.ForgetDeviceRequest{IdempotencyKey: key, MemberId: id, ApnsToken: token}))
		return err
	}
	for name, call := range map[string]func() error{
		"no key":         func() error { return forget(as("ann"), "", ann, "f0f0") },
		"another Member": func() error { return forget(as("bob"), "k9", ann, "f0f0") },
	} {
		if err := call(); err == nil {
			t.Errorf("%s: forgotten, want refused", name)
		}
	}
	for _, key := range []string{"k1", "k1"} {
		if err := forget(as("ann"), key, ann, "F0F0"); err != nil {
			t.Fatalf("ann forgetting her phone under %s: %v", key, err)
		}
	}
	if err := forget(as("ann"), "k2", ann, "b0b0"); err != nil {
		t.Fatalf("ann forgetting bob's token: %v", err)
	}
	rows, err := s.DB.Query(t.Context(), `SELECT apns_token FROM device WHERE member_id = ANY($1)`, []string{ann, bob})
	if err != nil {
		t.Fatal(err)
	}
	left, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil || len(left) != 1 || left[0] != "b0b0" {
		t.Errorf("devices left %v (%v), want bob's alone", left, err)
	}
}
