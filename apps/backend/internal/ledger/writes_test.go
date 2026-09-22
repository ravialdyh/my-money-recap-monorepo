package ledger

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// fakeMembers answers GetProfile for the Members each bearer token owns, as members does (§4.2 point 3); the
// tests' tokens are the subject itself.
type fakeMembers struct {
	moneyrecapv1connect.UnimplementedMembersServiceHandler
	owns map[string]*moneyrecapv1.Profile
}

func (f fakeMembers) GetProfile(_ context.Context, req *connect.Request[moneyrecapv1.GetProfileRequest]) (*connect.Response[moneyrecapv1.GetProfileResponse], error) {
	p, ok := f.owns[strings.TrimPrefix(req.Header().Get("Authorization"), "Bearer ")]
	if !ok || p.MemberId != req.Msg.MemberId {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	return connect.NewResponse(&moneyrecapv1.GetProfileResponse{Profile: p}), nil
}

func withMembers(t *testing.T, s *Service, owns map[string]*moneyrecapv1.Profile) {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewMembersServiceHandler(fakeMembers{owns: owns}))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	s.Members = moneyrecapv1connect.NewMembersServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(auth.Forward()))
}

func idr(amount string) *moneyrecapv1.Money {
	return &moneyrecapv1.Money{Amount: amount, CurrencyCode: "IDR"}
}

// The first Salary and the first Target take effect in the sign-up month (spec #92); later rows follow D56
// and D65; two rows sharing a month resolve newest first; a key replays, and a reused key is refused.
func TestSalaryAndTargetRows(t *testing.T) {
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a001"
	ctx := seed(t, db, id, "rows", "Asia/Jakarta", "2026-09-10", "IDR", nil, nil)
	clock := at(t, "2026-09-20", "Asia/Jakarta")
	s := &Service{DB: db, Now: func() time.Time { return clock }}
	salary := func(key, amount, payday string, applies moneyrecapv1.SalaryAppliesFrom) (*moneyrecapv1.SaveSalaryResponse, error) {
		res, err := s.SaveSalary(ctx, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{IdempotencyKey: key, MemberId: id, Amount: idr(amount), Payday: payday, AppliesFrom: applies}))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	rows := func(table string) (n int) {
		t.Helper()
		if err := db.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE member_id = $1`, id).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	this := moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_THIS_MONTH

	if _, err := s.SaveTarget(ctx, connect.NewRequest(&moneyrecapv1.SaveTargetRequest{IdempotencyKey: "t0", MemberId: id, Percent: 20})); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("a Target before any Salary: %v, want failed_precondition (D44)", err)
	}
	first, err := salary("s1", "3000000", "25", this)
	if err != nil || first.Settings.Salary.EffectiveFrom != "2026-09" || first.Month.Salary.DaysUntilPayday != 5 {
		t.Fatalf("first Salary: %v, %v, want effective 2026-09 and 5 days to the 25th", first, err)
	}
	again, err := salary("s1", "3000000", "25", this)
	if err != nil || !strings.Contains(again.String(), "3000000") || rows("salary_setting") != 1 {
		t.Fatalf("s1 again: %v, %v, %d rows; want the first answer and 1 row", again, err, rows("salary_setting"))
	}
	if _, err := salary("s1", "3100000", "25", this); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("s1 reused for another amount: %v, want invalid_argument", err)
	}
	newest, err := salary("s2", "3500000", "25", this)
	if err != nil || newest.Settings.Salary.Amount.Amount != "3500000" || newest.Settings.Salary.EffectiveFrom != "2026-09" {
		t.Fatalf("a second September row: %v, %v, want 3500000 in force from 2026-09, the newest", newest, err)
	}
	moved, err := salary("s3", "3500000", "last", this)
	if err != nil || moved.Settings.Salary.Payday != "25" || moved.Settings.NextSalary.GetPayday() != "last" || moved.Settings.NextSalary.GetEffectiveFrom() != "2026-10" {
		t.Fatalf("a new Payday: %v, %v, want the 25th this month and last from 2026-10 (D56)", moved, err)
	}
	clock = at(t, "2026-09-26", "Asia/Jakarta")
	if _, err := salary("s4", "3600000", "25", moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_UNSPECIFIED); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Errorf("unspecified after this month's Payday: %v, want failed_precondition (D56 sheet)", err)
	}
	for name, req := range map[string]*moneyrecapv1.SaveSalaryRequest{
		"no key":          {MemberId: id, Amount: idr("1"), Payday: "25"},
		"zero":            {IdempotencyKey: "b1", MemberId: id, Amount: idr("0"), Payday: "25"},
		"twelve digits":   {IdempotencyKey: "b2", MemberId: id, Amount: idr("100000000000"), Payday: "25"},
		"a rupiah cent":   {IdempotencyKey: "b3", MemberId: id, Amount: idr("3000000.5"), Payday: "25"},
		"another money":   {IdempotencyKey: "b4", MemberId: id, Amount: &moneyrecapv1.Money{Amount: "2400", CurrencyCode: "GBP"}, Payday: "25"},
		"payday 32":       {IdempotencyKey: "b5", MemberId: id, Amount: idr("1"), Payday: "32"},
		"no payday":       {IdempotencyKey: "b6", MemberId: id, Amount: idr("1")},
		"a grouped value": {IdempotencyKey: "b7", MemberId: id, Amount: idr("3.000.000"), Payday: "25"},
	} {
		if _, err := s.SaveSalary(ctx, connect.NewRequest(req)); connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v, want invalid_argument", name, err)
		}
	}
	target := func(key string, percent int32) (*moneyrecapv1.SaveTargetResponse, error) {
		res, err := s.SaveTarget(ctx, connect.NewRequest(&moneyrecapv1.SaveTargetRequest{IdempotencyKey: key, MemberId: id, Percent: percent}))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	firstTarget, err := target("t1", 20)
	if err != nil || firstTarget.Settings.SavingTarget.GetEffectiveFrom() != "2026-09" || firstTarget.Month.SavingTarget.GetAmount().GetAmount() != "700000" {
		t.Fatalf("first Target: %v, %v, want 20%% from 2026-09, 3500000 × 20%% = 700000", firstTarget, err)
	}
	later, err := target("t2", 30)
	if err != nil || later.Settings.SavingTarget.GetPercent() != 20 || later.Settings.NextSavingTarget.GetPercent() != 30 || later.Settings.NextSavingTarget.GetEffectiveFrom() != "2026-10" {
		t.Fatalf("a later Target: %v, %v, want 20%% now and 30%% from 2026-10 (D65)", later, err)
	}
	if _, err := target("t3", 101); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("101%%: %v, want invalid_argument", err)
	}
	if n := rows("saving_target_setting"); n != 2 {
		t.Errorf("%d target rows, want 2", n)
	}
}

// A Member whose setup ends the month after sign-up still gets a sign-up-month Salary (spec #92). The
// replica fills from members under the caller's token; a token not owning the Member gets not_found.
func TestTheReplicaIsFilledUnderTheCallersToken(t *testing.T) {
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a002"
	owner := auth.NewContext(context.Background(), auth.Claims{Subject: "owner", Raw: "owner"})
	stranger := auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})
	t.Cleanup(func() {
		for _, table := range []string{"member_facts", "salary_setting", "idempotency"} {
			if _, err := db.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = $1`, id); err != nil {
				t.Logf("cleanup: %v", err)
			}
		}
	})
	s := &Service{DB: db, Now: func() time.Time { return at(t, "2026-09-02", "Asia/Jakarta") }}
	profile := &moneyrecapv1.Profile{MemberId: id, Timezone: "Asia/Jakarta", SignedUpOn: "2026-08-30"}
	withMembers(t, s, map[string]*moneyrecapv1.Profile{"owner": profile})

	get := func(ctx context.Context) error {
		_, err := s.GetSettings(ctx, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: id}))
		return err
	}
	if err := get(owner); connect.CodeOf(err) != connect.CodeFailedPrecondition {
		t.Fatalf("a Member with no Currency yet: %v, want failed_precondition", err)
	}
	profile.CurrencyCode = "IDR"
	if err := get(stranger); connect.CodeOf(err) != connect.CodeNotFound {
		t.Fatalf("a token that does not own the Member, before any fill: %v, want not_found", err)
	}
	res, err := s.SaveSalary(owner, connect.NewRequest(&moneyrecapv1.SaveSalaryRequest{IdempotencyKey: "k1", MemberId: id, Amount: idr("3000000"), Payday: "25"}))
	if err != nil || res.Msg.Settings.GetSalary().GetEffectiveFrom() != "2026-08" {
		t.Fatalf("first Salary on 2 September for a 30 August sign-up: %v, %v, want effective 2026-08", res, err)
	}
	var subject string
	if err := db.QueryRow(context.Background(), `SELECT keycloak_subject FROM member_facts WHERE member_id = $1`, id).Scan(&subject); err != nil || subject != "owner" {
		t.Fatalf("replica subject %q (%v), want owner", subject, err)
	}
	if err := get(stranger); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("a token that does not own the Member, after the fill: %v, want not_found", err)
	}
	if _, err := s.GetSettings(owner, connect.NewRequest(&moneyrecapv1.GetSettingsRequest{MemberId: "not-a-uuid"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("a malformed id: %v, want not_found", err)
	}
}
