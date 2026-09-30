package ledger

import (
	"context"
	"math/big"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// A rate is a positive decimal of at most 11 digits before the separator and 12 after, to a real Currency (spec #169).
func TestARateHasLimits(t *testing.T) {
	for rate, ok := range map[string]bool{
		"20150": true, "0.000049600000": true, "99999999999.999999999999": true, "1": true,
		"0": false, "0.000000000000": false, "999999999999": false, "0.0000000000001": false, "-1": false, "1e3": false, "1,5": false, "": false, ".5": false,
	} {
		if _, err := conversionOf(rate, "GBP"); (err == nil) != ok {
			t.Errorf("rate %q: %v, want ok %v", rate, err, ok)
		}
	}
	if _, err := conversionOf("1", "XXX"); err == nil {
		t.Error("a code that is no Currency passed")
	}
}

// ApplyCurrencyConversion replays its key and refuses it reused; a D57 overflow is out_of_range in both calls with
// nothing written; another Member's amounts, the Currency held now and a missing source are refused.
func TestApplyCurrencyConversion(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000c173"
	ctx, s, log := spendsOf(t, id, "converter")
	log("k1", "150000", "2026-09-10")
	apply := func(key, rate, to string, source moneyrecapv1.RateSource) (*moneyrecapv1.ApplyCurrencyConversionResponse, error) {
		res, err := s.ApplyCurrencyConversion(ctx, connect.NewRequest(&moneyrecapv1.ApplyCurrencyConversionRequest{IdempotencyKey: key, MemberId: id, Rate: rate, Source: source, ToCurrencyCode: to}))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	manual := moneyrecapv1.RateSource_RATE_SOURCE_MANUAL
	if _, err := s.PreviewConversion(ctx, connect.NewRequest(&moneyrecapv1.PreviewConversionRequest{MemberId: id, Rate: "40000", ToCurrencyCode: "USD"})); connect.CodeOf(err) != connect.CodeOutOfRange {
		t.Errorf("a preview past 11 digits: %v, want out_of_range", err)
	}
	for name, err := range map[string]error{
		"past 11 digits":         second(apply("big", "40000", "USD", manual)),
		"the Currency held now":  second(apply("same", "1", "IDR", manual)),
		"no source":              second(apply("nosource", "0.0001", "USD", moneyrecapv1.RateSource_RATE_SOURCE_UNSPECIFIED)),
		"no key":                 second(apply("", "0.0001", "USD", manual)),
		"a rate of 13 decimals":  second(apply("long", "0.0000000000001", "USD", manual)),
		"a code that is no code": second(apply("code", "1", "XXX", manual)),
	} {
		if want := map[bool]connect.Code{true: connect.CodeOutOfRange, false: connect.CodeInvalidArgument}[name == "past 11 digits"]; connect.CodeOf(err) != want {
			t.Errorf("%s: %v, want %v", name, err, want)
		}
	}
	var rows int
	if err := s.DB.QueryRow(ctx, `SELECT (SELECT count(*) FROM currency_conversion WHERE member_id = $1) + (SELECT count(*) FROM outbox WHERE member_id = $1 AND routing_key = 'ledger.currency.converted')`, id).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("after the refusals: %d conversion and event rows (%v), want none", rows, err)
	}

	otherID := "00000000-0000-4000-8000-00000000c174"
	otherCtx, _, otherLog := spendsOf(t, otherID, "other-converter")
	otherLog("o1", "1000", "2026-09-20")
	for name, c := range map[string]context.Context{"the caller's token": ctx, "a stranger's token": auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})} {
		if _, err := s.ApplyCurrencyConversion(c, connect.NewRequest(&moneyrecapv1.ApplyCurrencyConversionRequest{
			IdempotencyKey: "theirs", MemberId: otherID, Rate: "0.0001", Source: manual, ToCurrencyCode: "USD",
		})); connect.CodeOf(err) != connect.CodeNotFound {
			t.Errorf("another Member's amounts under %s: %v, want not_found", name, err)
		}
	}

	first, err := apply("c1", "0.0000625", "USD", manual)
	if err != nil {
		t.Fatal(err)
	}
	if first.Settings.CurrencyCode != "USD" || first.Settings.Salary.GetAmount().GetAmount() != "187.50" || first.Month.Spent.GetAmount() != "9.38" {
		t.Fatalf("the conversion answered %v, want USD, a 187.50 Salary and 9.38 spent (150,000 × 0.0000625 = 9.375, half up)", first)
	}
	again, err := apply("c1", "0.0000625", "USD", manual)
	if err != nil || !proto.Equal(again, first) {
		t.Fatalf("the same key again: %v, %v; want the first answer", again, err)
	}
	if _, err := apply("c1", "0.00007", "USD", manual); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("the key reused for another rate: %v, want invalid_argument", err)
	}
	if err := s.DB.QueryRow(ctx, `SELECT (SELECT count(*) FROM currency_conversion WHERE member_id = $1) * 10 + (SELECT count(*) FROM outbox WHERE member_id = $1 AND routing_key = 'ledger.currency.converted')`, id).Scan(&rows); err != nil || rows != 11 {
		t.Errorf("after one conversion and its replay: %d (conversion rows × 10 + events) (%v), want 11", rows, err)
	}
	res, err := s.ListSpends(otherCtx, connect.NewRequest(&moneyrecapv1.ListSpendsRequest{MemberId: otherID, Month: "2026-09"}))
	if err != nil || len(res.Msg.Spends) != 1 || res.Msg.Spends[0].Amount.Amount != "1000" || res.Msg.Spends[0].Amount.CurrencyCode != "IDR" {
		t.Errorf("the other Member's September: %v, %v, want their Spend untouched", res, err)
	}
}

// The conversion stamps the replica's Currency, so a members event from before it cannot take the code back.
func TestAConversionOutlivesALateMembersEvent(t *testing.T) {
	id := "00000000-0000-4000-8000-00000000c175"
	ctx, s, _ := spendsOf(t, id, "late-event")
	if _, err := s.ApplyCurrencyConversion(ctx, connect.NewRequest(&moneyrecapv1.ApplyCurrencyConversionRequest{
		IdempotencyKey: "c", MemberId: id, Rate: "0.0000625", Source: moneyrecapv1.RateSource_RATE_SOURCE_PROVIDER, ToCurrencyCode: "USD",
	})); err != nil {
		t.Fatal(err)
	}
	late := &moneyrecapv1.MembersUpdated{Envelope: envelope(id, time.Now().Add(-time.Minute)), Facts: &moneyrecapv1.MemberFacts{CurrencyCode: "IDR"}}
	if err := apply(t, s.DB, late); err != nil {
		t.Fatal(err)
	}
	if got := replica(t, s.DB, id); got != "late-event Asia/Jakarta 2026-07-01 USD" {
		t.Fatalf("after a members event older than the conversion the replica is %q, want USD kept", got)
	}
}

// A rewrite rounds half up to the new decimals and, as the relabel does, never leaves an amount at zero.
func TestARewriteRoundsHalfUpAndStaysAboveZero(t *testing.T) {
	c, err := conversionOf("0.00002", "BHD")
	if err != nil {
		t.Fatal(err)
	}
	for amount, want := range map[string]string{"1": "0.001", "25": "0.001", "75": "0.002", "1000000": "20.000"} {
		r, _ := new(big.Rat).SetString(amount)
		if got := c.rewritten(r).FloatString(3); got != want {
			t.Errorf("Rp%s at 0.00002: %s BHD, want %s", amount, got, want)
		}
	}
}

func second[T any](_ T, err error) error { return err }
