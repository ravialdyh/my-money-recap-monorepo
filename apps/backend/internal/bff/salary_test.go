package bff

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

func keyed[T any](msg *T, key string) *connect.Request[T] {
	req := connect.NewRequest(msg)
	if key != "" {
		req.Header().Set("Idempotency-Key", key)
	}
	return req
}

func salaryAt(step int32) *moneyrecapv1.Profile {
	return &moneyrecapv1.Profile{MemberId: "m-1", DisplayName: "Maya", AvatarKey: "fox", CurrencyCode: "IDR", OnboardingStep: step}
}

// During setup ledger saves the Salary as a setup save, then members records step 3 under the same key; a failed
// ledger write records no step, a re-save records none again, and after setup the save is an edit (spec #92).
func TestSaveSalaryWritesLedgerThenTheStep(t *testing.T) {
	this := moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_THIS_MONTH
	msg := &moneyrecapv1.BffServiceSaveSalaryRequest{Amount: &moneyrecapv1.Money{Amount: "15000000", CurrencyCode: "IDR"}, Payday: "25", AppliesFrom: this}
	members, ledger := &fakeMembers{profile: salaryAt(2)}, &fakeLedger{}
	s := newService(t, members, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	if _, err := s.SaveSalary(ctx, keyed(msg, "")); connect.CodeOf(err) != connect.CodeInvalidArgument || ledger.gotSalary != nil {
		t.Fatalf("no key: %v, ledger reached %v; want invalid_argument before any call", err, ledger.gotSalary != nil)
	}
	res, err := s.SaveSalary(ctx, keyed(msg, "k1"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ledger.gotSalary; !got.Setup || got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.AppliesFrom != this || ledger.gotToken != "Bearer tok" {
		t.Errorf("ledger got %v with %q, want a setup save under k1 for m-1 with the caller's token", got, ledger.gotToken)
	}
	if got := members.gotProfile; got.GetIdempotencyKey() != "k1" || got.GetOnboardingStep() != 3 || got.DisplayName != nil || got.CurrencyCode != nil {
		t.Errorf("members got %v, want step 3 alone under k1", got)
	}
	if v := res.Msg; v.Settings.Profile.OnboardingStep != 3 || v.Settings.LedgerSettings.Salary.GetPayday() != "25" || v.Home.CurrencyCode != "IDR" {
		t.Errorf("answer %v, want step 3, the Salary and a Home in IDR", v)
	}

	members.gotProfile = nil
	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeInvalidArgument, errLedgerDown)})
	if _, err := s.SaveSalary(ctx, keyed(msg, "k2")); connect.CodeOf(err) != connect.CodeInvalidArgument || members.gotProfile != nil {
		t.Errorf("a refused Salary: %v, step recorded %v; want invalid_argument and no step", err, members.gotProfile != nil)
	}
	for step, setup := range map[int32]bool{3: true, 4: false} {
		members, ledger := &fakeMembers{profile: salaryAt(step)}, &fakeLedger{}
		s := newService(t, members, &fakeNotify{})
		s.Ledger = ledgerClient(t, ledger)
		if _, err := s.SaveSalary(ctx, keyed(msg, "k3")); err != nil || ledger.gotSalary.Setup != setup || members.gotProfile != nil {
			t.Errorf("at step %d: %v, setup %v, step recorded %v; want setup %v and no step write", step, err, ledger.gotSalary.GetSetup(), members.gotProfile != nil, setup)
		}
	}
}

// Step 4 mirrors step 3: a setup save in ledger, then step 4 in members under the same key; D44's refusal passes
// through with no step recorded; after setup the save is an edit and no step is sent (spec #92, owner 2026-09-23).
func TestSaveTargetWritesLedgerThenTheStep(t *testing.T) {
	msg := &moneyrecapv1.BffServiceSaveTargetRequest{Percent: 20}
	members, ledger := &fakeMembers{profile: salaryAt(3)}, &fakeLedger{}
	s := newService(t, members, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	if _, err := s.SaveTarget(ctx, keyed(msg, "")); connect.CodeOf(err) != connect.CodeInvalidArgument || ledger.gotTarget != nil {
		t.Fatalf("no key: %v, ledger reached %v; want invalid_argument before any call", err, ledger.gotTarget != nil)
	}
	res, err := s.SaveTarget(ctx, keyed(msg, "k1"))
	if err != nil {
		t.Fatal(err)
	}
	if got := ledger.gotTarget; !got.Setup || got.IdempotencyKey != "k1" || got.MemberId != "m-1" || got.Percent != 20 || ledger.gotToken != "Bearer tok" {
		t.Errorf("ledger got %v with %q, want a setup save of 20%% under k1 for m-1 with the caller's token", got, ledger.gotToken)
	}
	if got := members.gotProfile; got.GetIdempotencyKey() != "k1" || got.GetOnboardingStep() != 4 {
		t.Errorf("members got %v, want step 4 alone under k1", got)
	}
	if v := res.Msg; v.Settings.Profile.OnboardingStep != 4 || v.Settings.LedgerSettings.SavingTarget.GetPercent() != 20 {
		t.Errorf("answer %v, want step 4 and the Target", v)
	}

	blocked := &fakeMembers{profile: salaryAt(3)}
	s = newService(t, blocked, &fakeNotify{})
	s.Ledger = ledgerClient(t, &fakeLedger{err: connect.NewError(connect.CodeFailedPrecondition, errors.New("no Saving Target without a Salary (D44)"))})
	if _, err := s.SaveTarget(ctx, keyed(msg, "k2")); connect.CodeOf(err) != connect.CodeFailedPrecondition || blocked.gotProfile != nil {
		t.Errorf("a Target with no Salary: %v, step recorded %v; want failed_precondition and no step", err, blocked.gotProfile != nil)
	}
	done, edit := &fakeMembers{profile: salaryAt(4)}, &fakeLedger{}
	s = newService(t, done, &fakeNotify{})
	s.Ledger = ledgerClient(t, edit)
	if _, err := s.SaveTarget(ctx, keyed(msg, "k3")); err != nil || edit.gotTarget.GetSetup() || done.gotProfile != nil {
		t.Errorf("after setup: %v, setup %v, step recorded %v; want an edit and no step write", err, edit.gotTarget.GetSetup(), done.gotProfile != nil)
	}
}

// A Currency saved again during setup relabels step 3's Salary in ledger, after members holds it, and the answer
// carries the relabelled Salary; a first Currency moves nothing; after setup a change is refused (spec #92).
func TestSaveProfileRelabelsDuringSetupAndRefusesAfter(t *testing.T) {
	change := &moneyrecapv1.BffServiceSaveProfileRequest{CurrencyCode: proto.String("EUR"), OnboardingStep: proto.Int32(2)}
	members, ledger := &fakeMembers{profile: salaryAt(3)}, &fakeLedger{salary: &moneyrecapv1.SalarySetting{Amount: &moneyrecapv1.Money{Amount: "15000000", CurrencyCode: "IDR"}}}
	s := newService(t, members, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	res, err := s.SaveProfile(ctx, keyed(change, "k1"))
	if err != nil || members.gotProfile.GetCurrencyCode() != "EUR" || ledger.gotRelabel.GetCurrencyCode() != "EUR" || ledger.gotRelabel.GetIdempotencyKey() != "k1" {
		t.Fatalf("back at step 2: %v, members %v, ledger %v; want EUR in both under k1", err, members.gotProfile, ledger.gotRelabel)
	}
	if got := res.Msg.Salary; got.GetAmount().GetCurrencyCode() != "EUR" || got.GetAmount().GetAmount() != "15000000" {
		t.Errorf("answered Salary %v, want 15000000 relabelled EUR", got)
	}

	first := &fakeLedger{}
	s = newService(t, &fakeMembers{profile: &moneyrecapv1.Profile{MemberId: "m-1", DisplayName: "Maya", AvatarKey: "fox", OnboardingStep: 1}}, &fakeNotify{})
	s.Ledger = ledgerClient(t, first)
	if _, err := s.SaveProfile(ctx, keyed(change, "k2")); err != nil || first.gotRelabel != nil {
		t.Errorf("a first Currency: %v, relabel sent %v; want saved with nothing to move", err, first.gotRelabel)
	}

	done := &fakeMembers{profile: salaryAt(4)}
	s = newService(t, done, &fakeNotify{})
	if _, err := s.SaveProfile(ctx, keyed(change, "k3")); connect.CodeOf(err) != connect.CodeFailedPrecondition || done.gotProfile != nil {
		t.Errorf("a change after setup: %v, members reached %v; want failed_precondition before any write", err, done.gotProfile != nil)
	}
	same := &moneyrecapv1.BffServiceSaveProfileRequest{CurrencyCode: proto.String("IDR")}
	if _, err := s.SaveProfile(ctx, keyed(same, "k4")); err != nil {
		t.Errorf("the same Currency after setup: %v, want it accepted", err)
	}
}

// A resume at step 3 or 4, or Home's prompts after setup, carry the saved Salary and Target from ledger; before the
// Currency ledger is never asked, and a ledger that fails is unavailable rather than an empty step (spec #92).
func TestGetOnboardingStateCarriesStepThreesAnswer(t *testing.T) {
	saved := &moneyrecapv1.SalarySetting{Amount: &moneyrecapv1.Money{Amount: "15000000", CurrencyCode: "IDR"}, Payday: "last"}
	target := &moneyrecapv1.SavingTargetSetting{Percent: 35, EffectiveFrom: "2026-09"}
	ledger := &fakeLedger{salary: saved, target: target}
	s := newService(t, &fakeMembers{profile: salaryAt(2)}, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	res, err := s.GetOnboardingState(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetOnboardingStateRequest{}))
	if err != nil || res.Msg.State.NextStep != 3 || !proto.Equal(res.Msg.State.Salary, saved) || !proto.Equal(res.Msg.State.SavingTarget, target) {
		t.Fatalf("at step 3: %v, %v; want step 3 with the saved Salary and Target", res, err)
	}
	s = newService(t, &fakeMembers{profile: salaryAt(4)}, &fakeNotify{})
	s.Ledger = ledgerClient(t, ledger)
	res, err = s.GetOnboardingState(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetOnboardingStateRequest{}))
	if err != nil || res.Msg.State.NextStep != 0 || !proto.Equal(res.Msg.State.Salary, saved) || !proto.Equal(res.Msg.State.SavingTarget, target) {
		t.Fatalf("after setup: %v, %v; want setup over with the Salary step 4 takes a share of", res, err)
	}
	early := &fakeLedger{err: errLedgerDown}
	s = newService(t, &fakeMembers{profile: &moneyrecapv1.Profile{MemberId: "m-1"}}, &fakeNotify{})
	s.Ledger = ledgerClient(t, early)
	if _, err := s.GetOnboardingState(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetOnboardingStateRequest{})); err != nil || early.settings != 0 {
		t.Errorf("at step 1: %v, ledger asked %d times; want no ledger call", err, early.settings)
	}
	s = newService(t, &fakeMembers{profile: salaryAt(3)}, &fakeNotify{})
	s.Ledger = ledgerClient(t, &fakeLedger{err: errLedgerDown})
	if _, err := s.GetOnboardingState(ctx, connect.NewRequest(&moneyrecapv1.BffServiceGetOnboardingStateRequest{})); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Errorf("a ledger that fails at step 4: %v, want unavailable, which the phone reads as no connection", err)
	}
}
