package ledger

import (
	"context"
	"errors"
	"math/big"
	"regexp"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
)

var (
	amountPattern = regexp.MustCompile(`^\d{1,11}(\.\d+)?$`)
	paydayPattern = regexp.MustCompile(`^([1-9]|[12]\d|3[01])$`)
)

func (s *Service) SaveSalary(ctx context.Context, req *connect.Request[moneyrecapv1.SaveSalaryRequest]) (*connect.Response[moneyrecapv1.SaveSalaryResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	amount, err := l.amount(req.Msg.Amount)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	payday := req.Msg.Payday
	if payday != "last" && !paydayPattern.MatchString(payday) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New(`payday is 1 to 31 or "last"`))
	}
	from := monthOf(l.signedUp)
	if !req.Msg.Setup {
		if from, err = l.salaryStarts(payday, req.Msg.AppliesFrom); err != nil {
			return nil, err
		}
	}
	answer := &moneyrecapv1.SaveSalaryResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO salary_setting (member_id, amount, original_amount, original_currency, payday, effective_from) VALUES ($1, $2, $2, $3, $4, $5)`,
			l.id, amount.FloatString(l.cash.scale), l.cash.code, payday, from.day(1)); err != nil {
			return err
		}
		l.salaries = append(l.salaries, salaryRow{amount: amount, original: amount, originalCurrency: l.cash.code, payday: payday, from: from})
		answer.Settings, answer.Month = l.settings(l.today, req.Msg.Setup), l.summary(monthOf(l.today), l.today)
		return nil
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func (s *Service) RelabelCurrency(ctx context.Context, req *connect.Request[moneyrecapv1.RelabelCurrencyRequest]) (*connect.Response[moneyrecapv1.RelabelCurrencyResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	c, err := currencyOf(req.Msg.CurrencyCode)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	unit := c.unit().FloatString(c.scale)
	answer := &moneyrecapv1.RelabelCurrencyResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE member_facts SET currency_code = $2, currency_at = now() WHERE member_id = $1`, l.id, c.code); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE salary_setting SET original_currency = $2,
			amount = greatest(round(amount, $3::int), $4::numeric), original_amount = greatest(round(original_amount, $3::int), $4::numeric)
			WHERE member_id = $1`, l.id, c.code, c.scale, unit); err != nil {
			return err
		}
		if l.salaries, err = salaries(ctx, tx, l.id); err != nil {
			return err
		}
		l.cash = c
		answer.Settings = l.settings(l.today, false)
		return nil
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func (l ledgerOf) amount(m *moneyrecapv1.Money) (*big.Rat, error) {
	if m.GetCurrencyCode() != l.cash.code || !amountPattern.MatchString(m.GetAmount()) {
		return nil, errors.New("amount is not a decimal in the Member's currency")
	}
	r, _ := new(big.Rat).SetString(m.GetAmount())
	if r.Sign() <= 0 || l.cash.rounded(r).Cmp(r) != 0 {
		return nil, errors.New("amount is not above zero with the currency's decimals")
	}
	return r, nil
}

func (l member) salaryStarts(payday string, applies moneyrecapv1.SalaryAppliesFrom) (month, error) {
	now := monthOf(l.today)
	current, ok := inForce(l.salaries, salaryFrom, now)
	switch {
	case len(l.salaries) == 0:
		return now, nil
	case ok && current.payday != payday, applies == moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_NEXT_MONTH:
		return now.next(), nil
	case applies == moneyrecapv1.SalaryAppliesFrom_SALARY_APPLIES_FROM_UNSPECIFIED && ok && !l.today.Before(now.day(effectivePayday(current.payday, now))):
		return month{}, connect.NewError(connect.CodeFailedPrecondition, errors.New("this month's Payday has come: say which month the Salary applies from"))
	}
	return now, nil
}

func (s *Service) SaveTarget(ctx context.Context, req *connect.Request[moneyrecapv1.SaveTargetRequest]) (*connect.Response[moneyrecapv1.SaveTargetResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	if req.Msg.Percent < 0 || req.Msg.Percent > 100 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("percent is 0 to 100"))
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	if len(l.salaries) == 0 {
		return nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("no Saving Target without a Salary"))
	}
	from := l.targetStarts(l.today, req.Msg.Setup)
	answer := &moneyrecapv1.SaveTargetResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO saving_target_setting (member_id, percent, effective_from) VALUES ($1, $2, $3)`, l.id, req.Msg.Percent, from.day(1)); err != nil {
			return err
		}
		l.targets = append(l.targets, targetRow{percent: req.Msg.Percent, from: from})
		answer.Settings, answer.Month = l.settings(l.today, req.Msg.Setup), l.summary(monthOf(l.today), l.today)
		return nil
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func writeFailed(ctx context.Context, err error) error {
	var ce *connect.Error
	if errors.As(err, &ce) {
		return ce
	}
	if errors.Is(err, idempotency.ErrReused) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	return storeFailed(ctx, err)
}
