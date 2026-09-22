package members

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// The onboarding step, the sign-up date and the deletion stamp reach the caller through both reads, and a
// Member who has none of them reads zero values rather than an error (#94).
func TestBothReadsReturnTheOnboardingAndDeletionColumns(t *testing.T) {
	db := testDB(t)
	subject := "columns-" + t.Name()
	ctx := auth.NewContext(context.Background(), auth.Claims{Subject: subject, Email: "columns@example.com"})
	t.Cleanup(func() {
		if _, err := db.Exec(context.Background(), `DELETE FROM member WHERE keycloak_subject = $1`, subject); err != nil {
			t.Logf("cleanup: %v", err)
		}
	})
	s := &Service{DB: db}
	reads := func() map[string]*moneyrecapv1.Profile {
		t.Helper()
		created, err := s.GetOrCreateBySubject(ctx, connect.NewRequest(&moneyrecapv1.GetOrCreateBySubjectRequest{Subject: subject, Email: "columns@example.com"}))
		if err != nil {
			t.Fatal(err)
		}
		direct, err := s.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: created.Msg.Profile.MemberId}))
		if err != nil {
			t.Fatal(err)
		}
		return map[string]*moneyrecapv1.Profile{"GetOrCreateBySubject": created.Msg.Profile, "GetProfile": direct.Msg.Profile}
	}
	for name, p := range reads() {
		if p.OnboardingStep != 0 || p.SignedUpOn != "" || p.DeletionRequestedAt != nil {
			t.Errorf("%s of a new Member: step %d, signed up %q, deletion %v, want 0, empty and absent", name, p.OnboardingStep, p.SignedUpOn, p.DeletionRequestedAt)
		}
	}
	asked := time.Date(2026, 9, 23, 1, 2, 3, 0, time.UTC)
	if _, err := db.Exec(ctx, `UPDATE member SET onboarding_step = 3, signed_up_on = '2026-09-22', deletion_requested_at = $2 WHERE keycloak_subject = $1`, subject, asked); err != nil {
		t.Fatal(err)
	}
	for name, p := range reads() {
		if p.OnboardingStep != 3 || p.SignedUpOn != "2026-09-22" || p.DeletionRequestedAt == nil || !p.DeletionRequestedAt.AsTime().Equal(asked) {
			t.Errorf("%s: step %d, signed up %q, deletion %v, want 3, 2026-09-22 and %v", name, p.OnboardingStep, p.SignedUpOn, p.DeletionRequestedAt, asked)
		}
	}
}
