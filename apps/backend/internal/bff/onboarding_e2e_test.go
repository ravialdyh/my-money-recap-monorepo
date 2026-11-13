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

// TestOnboardingEndToEnd walks steps 1 to 3 through the BFF, members and ledger containers as the realm's test
// Member: Home refused until the Currency, bad answers refused, Home greeting, back to step 1, then step 3. Skips
// like TestEndToEnd.
func TestOnboardingEndToEnd(t *testing.T) {
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
	zone := connect.NewRequest(&moneyrecapv1.BffServiceSetTimezoneRequest{Timezone: "Asia/Jakarta"})
	zone.Header().Set("Idempotency-Key", rand.Text())
	if _, err := client.SetTimezone(ctx, zone); err != nil {
		t.Fatal(err)
	}
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
	if _, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("Home after step 1: %v, want failed_precondition until the Currency", err)
	}
	if err := save(&moneyrecapv1.BffServiceSaveProfileRequest{CurrencyCode: proto.String("ZZZ"), OnboardingStep: proto.Int32(2)}); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a code that is no currency: %v, want invalid_argument", err)
	}
	if err := save(&moneyrecapv1.BffServiceSaveProfileRequest{CurrencyCode: proto.String("IDR"), OnboardingStep: proto.Int32(2)}); err != nil {
		t.Fatal(err)
	}
	if s := state(); s.NextStep != 3 || s.Profile.CurrencyCode != "IDR" || s.Profile.OnboardingStep != 2 {
		t.Fatalf("after step 2: %v, want step 3 with IDR", s)
	}
	home, err := client.GetHome(ctx, connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	assertZeroHome(t, home.Msg.Home, "Test", "IDR")
	if home.Msg.Home.AvatarKey != "owl" {
		t.Fatalf("Home after step 2 shows the %q Avatar, want owl", home.Msg.Home.AvatarKey)
	}
	t.Logf("steps 1 and 2 saved through the containers; Home greets %q with the %s in %s", home.Msg.Home.DisplayName, home.Msg.Home.AvatarKey, home.Msg.Home.CurrencyCode)

	if err := save(&moneyrecapv1.BffServiceSaveProfileRequest{DisplayName: proto.String("Tess"), AvatarKey: proto.String("fox"), OnboardingStep: proto.Int32(1)}); err != nil {
		t.Fatal(err)
	}
	if s := state(); s.NextStep != 3 || s.Profile.DisplayName != "Tess" || s.Profile.AvatarKey != "fox" || s.Profile.CurrencyCode != "IDR" || s.Profile.OnboardingStep != 2 {
		t.Fatalf("after going back to step 1: %v, want Tess and fox with IDR kept and progress still at step 2 (#102)", s)
	}
	t.Log("going back to step 1 changed only its answers: the Currency and the saved progress stayed")
	stepThree(ctx, t, client, db, testauth.Decode(t, tok).Sub, state, save)
	stepFour(ctx, t, client, state)
}

// stepThree saves step 3 through the BFF and ledger (#103): the Salary, then the step; a crash between the two,
// replayed by lowering the step; a new Payday replacing the Salary from the sign-up month; a Currency relabelling it.
func stepThree(ctx context.Context, t *testing.T, client moneyrecapv1connect.BffServiceClient, db *pgxpool.Pool, sub string, state func() *moneyrecapv1.OnboardingState, save func(*moneyrecapv1.BffServiceSaveProfileRequest) error) {
	t.Helper()
	salary := func(amount, payday string) error {
		req := connect.NewRequest(&moneyrecapv1.BffServiceSaveSalaryRequest{
			Amount: &moneyrecapv1.Money{Amount: amount, CurrencyCode: "IDR"}, Payday: payday,
			AppliesFrom: moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_THIS_MONTH,
		})
		req.Header().Set("Idempotency-Key", rand.Text())
		_, err := client.SaveSalary(ctx, req)
		return err
	}
	if err := salary("150000000.5", "25"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("a rupiah cent: %v, want invalid_argument, the Currency has 0 decimals", err)
	}
	if err := salary("150000000", "25"); err != nil {
		t.Fatal(err)
	}
	s := state()
	signUpMonth := s.Profile.SignedUpOn[:7]
	if s.NextStep != 4 || s.Profile.OnboardingStep != 3 || s.Salary.GetAmount().GetAmount() != "150000000" || s.Salary.GetEffectiveFrom() != signUpMonth {
		t.Fatalf("after step 3: %v, want step 4 next with 150000000 from %s", s, signUpMonth)
	}
	if _, err := db.Exec(ctx, `UPDATE member SET onboarding_step = 2 WHERE keycloak_subject = $1`, sub); err != nil {
		t.Fatal(err)
	}
	if s := state(); s.NextStep != 3 || s.Salary.GetAmount().GetAmount() != "150000000" {
		t.Fatalf("after a crash between the two writes: %v, want step 3 with the amount filled in", s)
	}
	if err := salary("150000000", "last"); err != nil {
		t.Fatal(err)
	}
	if s := state(); s.NextStep != 4 || s.Salary.GetPayday() != "last" || s.Salary.GetEffectiveFrom() != signUpMonth {
		t.Fatalf("step 3 again with a new Payday: %v, want last in force from %s, one Salary", s, signUpMonth)
	}
	if err := save(&moneyrecapv1.BffServiceSaveProfileRequest{CurrencyCode: proto.String("EUR"), OnboardingStep: proto.Int32(2)}); err != nil {
		t.Fatal(err)
	}
	if s := state(); s.Profile.CurrencyDigits != 2 || s.Salary.GetAmount().GetAmount() != "150000000.00" || s.Salary.GetOriginalAmount().GetCurrencyCode() != "EUR" {
		t.Fatalf("back at step 2 for EUR: %v, want 150000000.00 EUR, original too, and 2 decimals to type", s)
	}
	if err := save(&moneyrecapv1.BffServiceSaveProfileRequest{CurrencyCode: proto.String("IDR"), OnboardingStep: proto.Int32(2)}); err != nil {
		t.Fatal(err)
	}
	t.Log("step 3 saved, resumed after a crash, replaced with a new Payday, and relabelled EUR and back to IDR")
}

// stepFour saves step 4 through the BFF and ledger (#104): the Target from the sign-up month with its amount on Home,
// then setup over, so a second Target is an edit starting next month (D65).
func stepFour(ctx context.Context, t *testing.T, client moneyrecapv1connect.BffServiceClient, state func() *moneyrecapv1.OnboardingState) {
	t.Helper()
	target := func(percent int32) *moneyrecapv1.BffServiceSaveTargetResponse {
		t.Helper()
		req := connect.NewRequest(&moneyrecapv1.BffServiceSaveTargetRequest{Percent: percent})
		req.Header().Set("Idempotency-Key", rand.Text())
		res, err := client.SaveTarget(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg
	}
	first := target(20)
	signUpMonth := first.Settings.Profile.SignedUpOn[:7]
	if got := first.Home.SavingTarget; got.GetPercent() != 20 || got.GetAmount().GetAmount() != "30000000" || first.Settings.LedgerSettings.SavingTarget.GetEffectiveFrom() != signUpMonth {
		t.Fatalf("step 4 at 20%%: %v, want 30000000 of 150000000 on Home, from %s", first, signUpMonth)
	}
	if s := state(); s.NextStep != 0 || s.Profile.OnboardingStep != 4 || s.SavingTarget.GetPercent() != 20 || s.Salary.GetAmount().GetAmount() != "150000000" {
		t.Fatalf("after step 4: %v, want setup over with the Salary and Target Home's prompts reopen", s)
	}
	if edit := target(30); edit.Home.SavingTarget.GetPercent() != 20 || edit.Settings.LedgerSettings.NextSavingTarget.GetPercent() != 30 {
		t.Fatalf("a Target after setup: %v, want 20%% now and 30%% from next month (D65)", edit)
	}
	t.Log("step 4 saved from the sign-up month, setup over, and a later Target waits for next month")
}
