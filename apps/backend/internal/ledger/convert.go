package ledger

import (
	"context"
	"errors"
	"math/big"
	"regexp"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
)

// ratePattern is a rate as quoted or typed (spec #169): a decimal with at most 11 digits before the separator and
// 12 after.
var ratePattern = regexp.MustCompile(`^\d{1,11}(\.\d{1,12})?$`)

// errTooLong is D57 broken by a rewrite: the phone shows "That's more than the app can store." and nothing is written.
var errTooLong = errors.New("a converted amount would pass 11 digits before the separator (D57)")

// conversion is one Currency change: the Currency the amounts go into and the rate, units of it per current unit.
type conversion struct {
	to   cash
	rate *big.Rat
	text string
}

// conversionOf checks a rate and a code as sent: a positive ratePattern decimal and an ISO 4217 Currency.
func conversionOf(rate, code string) (conversion, error) {
	to, err := currencyOf(code)
	if err != nil {
		return conversion{}, err
	}
	r, ok := new(big.Rat).SetString(rate)
	if !ratePattern.MatchString(rate) || !ok || r.Sign() <= 0 {
		return conversion{}, errors.New("rate is not a positive decimal of at most 11 digits before the separator and 12 after")
	}
	return conversion{to: to, rate: r, text: rate}, nil
}

// rewritten is PRD §8.9 step 4 for one amount: amount × rate, rounded half up to the new decimals, never below the
// smallest unit, as the relabel keeps every amount above zero; the SQL rewrite in convert stores the same.
func (c conversion) rewritten(amount *big.Rat) *big.Rat {
	v := c.to.rounded(new(big.Rat).Mul(amount, c.rate))
	if v.Sign() == 0 {
		return c.to.unit()
	}
	return v
}

// preview is the confirm sheet's rows (PRD §8.9 step 3): the Salary in force and the total of every Spend, before
// and after. out_of_range when any Salary row or Spend would pass D57; invalid_argument for the Currency held now.
func (l member) preview(c conversion) (*moneyrecapv1.ConversionPreview, error) {
	if c.to.code == l.cash.code {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("the amounts are already in that currency"))
	}
	fits := func(r *big.Rat) bool { return amountPattern.MatchString(c.rewritten(r).FloatString(c.to.scale)) }
	for _, s := range l.salaries {
		if !fits(s.amount) {
			return nil, connect.NewError(connect.CodeOutOfRange, errTooLong)
		}
	}
	before, after := new(big.Rat), new(big.Rat)
	for _, s := range l.spends {
		if !fits(s.amount) {
			return nil, connect.NewError(connect.CodeOutOfRange, errTooLong)
		}
		before.Add(before, s.amount)
		after.Add(after, c.rewritten(s.amount))
	}
	p := &moneyrecapv1.ConversionPreview{
		FromCurrencyCode: l.cash.code, ToCurrencyCode: c.to.code, Rate: c.text, SpendsBefore: l.cash.money(before), SpendsAfter: c.to.money(after),
	}
	if sal, ok := inForce(l.salaries, salaryFrom, monthOf(l.today)); ok {
		p.SalaryBefore, p.SalaryAfter = l.cash.money(sal.amount), c.to.money(c.rewritten(sal.amount))
	}
	return p, nil
}

// PreviewConversion answers preview for the rate and Currency sent; nothing is written.
func (s *Service) PreviewConversion(ctx context.Context, req *connect.Request[moneyrecapv1.PreviewConversionRequest]) (*connect.Response[moneyrecapv1.PreviewConversionResponse], error) {
	c, err := conversionOf(req.Msg.Rate, req.Msg.ToCurrencyCode)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	p, err := l.preview(c)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.PreviewConversionResponse{Preview: p}), nil
}

// ApplyCurrencyConversion is PRD §8.9 step 4 in one transaction, checked as preview under the Member's lock: the rows
// rewritten, originals kept, the conversion row, the event, and the replica's Currency stamped as the relabel does.
func (s *Service) ApplyCurrencyConversion(ctx context.Context, req *connect.Request[moneyrecapv1.ApplyCurrencyConversionRequest]) (*connect.Response[moneyrecapv1.ApplyCurrencyConversionResponse], error) {
	msg := req.Msg
	if msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	if msg.Source != moneyrecapv1.RateSource_RATE_SOURCE_PROVIDER && msg.Source != moneyrecapv1.RateSource_RATE_SOURCE_MANUAL {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("source is PROVIDER or MANUAL"))
	}
	c, err := conversionOf(msg.Rate, msg.ToCurrencyCode)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	l, err := s.load(ctx, msg.MemberId)
	if err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.ApplyCurrencyConversionResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, msg.IdempotencyKey, msg, answer, func(tx pgx.Tx) error {
		if err := l.relock(ctx, tx); err != nil {
			return err
		}
		if l.salaries, err = salaries(ctx, tx, l.id); err != nil {
			return err
		}
		if _, err := l.preview(c); err != nil {
			return err
		}
		return l.convert(ctx, tx, c, msg.Source, answer)
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

// convert writes a checked conversion inside tx and fills answer with the settings and month in the new Currency.
func (l member) convert(ctx context.Context, tx pgx.Tx, c conversion, source moneyrecapv1.RateSource, answer *moneyrecapv1.ApplyCurrencyConversionResponse) (err error) {
	unit := c.to.unit().FloatString(c.to.scale)
	for _, table := range []string{"salary_setting", "spend"} {
		if _, err := tx.Exec(ctx, `UPDATE `+table+` SET amount = greatest(round(amount * $2::numeric, $3::int), $4::numeric) WHERE member_id = $1`,
			l.id, c.text, c.to.scale, unit); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `INSERT INTO currency_conversion (member_id, from_code, to_code, rate, source) VALUES ($1, $2, $3, $4, $5)`,
		l.id, l.cash.code, c.to.code, c.text, int16(source)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE member_facts SET currency_code = $2, currency_at = now() WHERE member_id = $1`, l.id, c.to.code); err != nil {
		return err
	}
	event := &moneyrecapv1.LedgerCurrencyConverted{FromCurrencyCode: l.cash.code, ToCurrencyCode: c.to.code, Rate: c.text, Source: source}
	if err := events.Stage(ctx, tx, l.id, "ledger.currency.converted", func(env *moneyrecapv1.EventEnvelope) proto.Message {
		event.Envelope = env
		return event
	}); err != nil {
		return err
	}
	l.cash = c.to
	if l.salaries, err = salaries(ctx, tx, l.id); err != nil {
		return err
	}
	if l.spends, err = spends(ctx, tx, l.id, l.cash); err != nil {
		return err
	}
	answer.Settings, answer.Month = l.settings(l.today, false), l.summary(monthOf(l.today), l.today)
	return nil
}
