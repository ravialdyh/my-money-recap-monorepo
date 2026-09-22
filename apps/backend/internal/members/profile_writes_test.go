package members

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// SaveProfile writes only the answers sent, stores the name in NFC, keeps the highest step when an earlier one
// is saved again, refuses what internal/answers refuses and honours the key (spec #92, A14).
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
