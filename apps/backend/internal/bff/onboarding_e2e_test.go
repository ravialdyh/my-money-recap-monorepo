package bff

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testauth"
)

// TestOnboardingStepOneEndToEnd walks step 1 through the BFF and members containers with the realm test
// Member's token: step 1 and its given name, the save, step 2, and Home greeting by name with the Avatar (#98).
// It skips like TestEndToEnd.
func TestOnboardingStepOneEndToEnd(t *testing.T) {
	base, kc, pw, dsn := os.Getenv("BFF_TEST_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("MMR_TEST_USER_PASSWORD"), os.Getenv("MMR_TEST_MEMBERS_DATABASE_URL")
	if base == "" || kc == "" || pw == "" || dsn == "" {
		t.Skip("BFF_TEST_URL, MMR_TEST_KEYCLOAK, MMR_TEST_USER_PASSWORD and MMR_TEST_MEMBERS_DATABASE_URL not all set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	tok := testauth.UserToken(t, kc, pw)
	if _, err := db.Exec(ctx, `DELETE FROM member WHERE keycloak_subject = $1`, testauth.Decode(t, tok).Sub); err != nil {
		t.Fatal(err)
	}
	client := moneyrecapv1connect.NewBffServiceClient(http.DefaultClient, base, testauth.Bearer(tok))
	state := func() *moneyrecapv1.OnboardingState {
		t.Helper()
		res, err := client.GetOnboardingState(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetOnboardingStateRequest{}))
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg.State
	}
	save := func(msg *moneyrecapv1.BffServiceSaveProfileRequest) error {
		req := connect.NewRequest(msg)
		req.Header().Set("Idempotency-Key", rand.Text())
		_, err := client.SaveProfile(ctx, req)
		return err
	}

	if s := state(); s.NextStep != 1 || s.SuggestedName != "Test" {
		t.Fatalf("a new Member's state: %v, want step 1 and the given name Test", s)
	}
	if err := save(&moneyrecapv1.BffServiceSaveProfileRequest{DisplayName: proto.String("Test"), AvatarKey: proto.String("cat"), OnboardingStep: proto.Int32(1)}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a seventh Avatar: %v, want invalid_argument", err)
	}
	if err := save(&moneyrecapv1.BffServiceSaveProfileRequest{DisplayName: proto.String("Test"), AvatarKey: proto.String("owl"), OnboardingStep: proto.Int32(1)}); err != nil {
		t.Fatal(err)
	}
	if s := state(); s.NextStep != 2 || s.Profile.DisplayName != "Test" || s.Profile.AvatarKey != "owl" || s.Profile.OnboardingStep != 1 {
		t.Fatalf("after step 1: %v, want step 2 with Test and owl", s)
	}
	home, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil || home.Msg.Home.DisplayName != "Test" || home.Msg.Home.AvatarKey != "owl" {
		t.Fatalf("Home after step 1: %v, %v; want the greeting's name and the Avatar", home, err)
	}
	t.Logf("step 1 saved through the containers; Home greets %q with the %s", home.Msg.Home.DisplayName, home.Msg.Home.AvatarKey)
}
