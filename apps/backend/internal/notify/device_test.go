package notify

import (
	"context"
	"testing"

	"connectrpc.com/connect"

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
