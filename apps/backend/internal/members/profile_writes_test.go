package members

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// SaveProfile writes only the answers sent, stores the name in NFC, keeps the highest step when an earlier one
// is saved again, answers the Currency's decimals, refuses what internal/answers refuses and honours the key (A14).
func TestSaveProfile(t *testing.T) {
	db := testDB(t)
	subject := "profile-" + t.Name()
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject})
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	m, err := Upsert(ctx, db, subject, "profile@example.com")
	if err != nil {
		t.Fatal(err)
	}
	s := &Service{DB: db}
	save := func(key string, req *moneyrecapv1.SaveProfileRequest) (*moneyrecapv1.Profile, error) {
		req.IdempotencyKey, req.MemberId = key, m.ID
		res, err := s.SaveProfile(ctx, connect.NewRequest(req))
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}
	want := func(step string, p *moneyrecapv1.Profile, err error, name, avatar, code string, n int32) {
		t.Helper()
		if err != nil || p.DisplayName != name || p.AvatarKey != avatar || p.CurrencyCode != code || p.OnboardingStep != n {
			t.Fatalf("%s: %v, %v; want %q %q %q step %d", step, p, err, name, avatar, code, n)
		}
	}
	p, err := save("k1", &moneyrecapv1.SaveProfileRequest{DisplayName: proto.String("  José "), AvatarKey: proto.String("otter"), OnboardingStep: proto.Int32(1)})
	want("step 1", p, err, "José", "otter", "", 1)
	p, err = save("k2", &moneyrecapv1.SaveProfileRequest{CurrencyCode: proto.String("IDR"), OnboardingStep: proto.Int32(2)})
	want("step 2", p, err, "José", "otter", "IDR", 2)
	p, err = save("k3", &moneyrecapv1.SaveProfileRequest{OnboardingStep: proto.Int32(3)})
	want("step 3 skipped", p, err, "José", "otter", "IDR", 3)
	p, err = save("k4", &moneyrecapv1.SaveProfileRequest{AvatarKey: proto.String("fox"), OnboardingStep: proto.Int32(1)})
	want("back to step 1", p, err, "José", "fox", "IDR", 3)
	p, err = save("k4", &moneyrecapv1.SaveProfileRequest{AvatarKey: proto.String("fox"), OnboardingStep: proto.Int32(1)})
	want("k4 again", p, err, "José", "fox", "IDR", 3)
	p, err = save("k5", &moneyrecapv1.SaveProfileRequest{CurrencyCode: proto.String("BHD"), OnboardingStep: proto.Int32(2)})
	want("back to step 2 for another Currency", p, err, "José", "fox", "BHD", 3)
	if p.CurrencyDigits != 3 {
		t.Errorf("BHD answers %d decimals, want 3, the CLDR data step 3 types with (spec #92)", p.CurrencyDigits)
	}
	if _, err := save("k4", &moneyrecapv1.SaveProfileRequest{AvatarKey: proto.String("owl")}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("k4 reused for another Avatar: %v, want invalid_argument", err)
	}
	for name, req := range map[string]*moneyrecapv1.SaveProfileRequest{
		"a digit":         {DisplayName: proto.String("Jose2")},
		"an emoji":        {DisplayName: proto.String("Jose 😀")},
		"31 letters":      {DisplayName: proto.String("abcdefghijklmnopqrstuvwxyzabcde")},
		"a seventh pet":   {AvatarKey: proto.String("cat")},
		"no currency":     {CurrencyCode: proto.String("ZZZ")},
		"step 5":          {OnboardingStep: proto.Int32(5)},
		"an empty answer": {DisplayName: proto.String("   ")},
	} {
		if _, err := save("bad-"+name, req); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", name, err)
		}
	}
	if _, err := s.SaveProfile(ctx, connect.NewRequest(&moneyrecapv1.SaveProfileRequest{MemberId: m.ID, AvatarKey: proto.String("owl")})); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no key: %v, want invalid_argument", err)
	}
	stored, err := Get(ctx, db, m.ID)
	if err != nil || stored.DisplayName != "José" || stored.AvatarKey != "fox" || stored.OnboardingStep != 3 {
		t.Fatalf("stored after the refused calls: %+v, %v", stored, err)
	}
}

// CancelDeletion clears a stamp written straight into the row and stages members.deletion.cancelled once; a
// replay, a Member with no stamp and a stamp past the 24 hours change nothing (D60, ARCHITECTURE §7.1 step 2).
func TestCancelDeletionClearsAStampUnderADayOld(t *testing.T) {
	db := testDB(t)
	subject := "cancel-" + t.Name()
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject})
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	m, err := Upsert(ctx, db, subject, "cancel@example.com")
	if err != nil {
		t.Fatal(err)
	}
	s := &Service{DB: db}
	stamp := func(ago time.Duration) {
		t.Helper()
		if _, err := db.Exec(ctx, `UPDATE member SET deletion_requested_at = now() - $2::interval WHERE id = $1`, m.ID, ago.String()); err != nil {
			t.Fatal(err)
		}
	}
	cancel := func(key string) (*moneyrecapv1.Profile, error) {
		res, err := s.CancelDeletion(ctx, connect.NewRequest(&moneyrecapv1.CancelDeletionRequest{IdempotencyKey: key, MemberId: m.ID}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Profile, nil
	}
	cancelled := func() int {
		n := 0
		for _, e := range outbox(t, db, m.ID) {
			if e.key == "members.deletion.cancelled" {
				n++
			}
		}
		return n
	}

	stamp(23 * time.Hour)
	if p, err := cancel("k1"); err != nil || p.DeletionRequestedAt != nil || p.MemberId != m.ID {
		t.Fatalf("23 hours old: %v, %v; want the Profile with no stamp", p, err)
	}
	if p, err := cancel("k1"); err != nil || p.DeletionRequestedAt != nil {
		t.Fatalf("k1 again: %v, %v", p, err)
	}
	if p, err := cancel("k2"); err != nil || p.DeletionRequestedAt != nil {
		t.Fatalf("no stamp: %v, %v", p, err)
	}
	if n := cancelled(); n != 1 {
		t.Errorf("%d members.deletion.cancelled staged, want 1", n)
	}
	stamp(25 * time.Hour)
	if _, err := cancel("k3"); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("25 hours old: %v, want failed_precondition", err)
	}
	if stored, err := Get(ctx, db, m.ID); err != nil || stored.DeletionRequestedAt == nil {
		t.Errorf("a stamp past the window was cleared: %+v, %v", stored, err)
	}
	if _, err := cancel(""); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("no key: %v, want invalid_argument", err)
	}
	if n := cancelled(); n != 1 {
		t.Errorf("%d members.deletion.cancelled staged after the refusals, want 1", n)
	}
}
