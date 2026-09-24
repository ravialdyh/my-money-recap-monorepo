package ledger

import (
	"context"
	"errors"
	"math/big"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"golang.org/x/text/unicode/norm"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
)

// LogSpend appends a Spend and answers it with its recomputed month and crossed_line (PRD §8.8). The Member's
// row is locked in the write's transaction and the Spends read again under it, so two logs cannot both miss the line.
func (s *Service) LogSpend(ctx context.Context, req *connect.Request[moneyrecapv1.LogSpendRequest]) (*connect.Response[moneyrecapv1.LogSpendResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	amount, date, note, err := l.spendFields(req.Msg.Amount, req.Msg.Category, req.Msg.Date, req.Msg.Note)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	answer := &moneyrecapv1.LogSpendResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SELECT 1 FROM member_facts WHERE member_id = $1 FOR UPDATE`, l.id); err != nil {
			return err
		}
		if l.spends, err = spends(ctx, tx, l.id, l.cash); err != nil {
			return err
		}
		m := monthOf(date)
		before := l.summary(m, l.today)
		spend := &moneyrecapv1.Spend{
			Amount: l.cash.money(amount), OriginalAmount: l.cash.money(amount), Category: req.Msg.Category, Date: req.Msg.Date, Note: note,
		}
		var created time.Time
		if err := tx.QueryRow(ctx, `INSERT INTO spend (member_id, amount, original_amount, original_currency, category, spent_on, note) VALUES ($1, $2, $2, $3, $4, $5, $6) RETURNING id::text, created_at`,
			l.id, spend.Amount.Amount, l.cash.code, int16(req.Msg.Category), date, note).Scan(&spend.Id, &created); err != nil {
			return err
		}
		spend.CreatedAt, spend.UpdatedAt = timestamppb.New(created), timestamppb.New(created)
		l.spends = append(l.spends, spendRow{amount: amount, date: date, spend: spend})
		answer.Spend, answer.Month = spend, l.summary(m, l.today)
		answer.CrossedLine = m == monthOf(l.today) && before.TargetWarning == nil && answer.Month.TargetWarning != nil
		return nil
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

// spendFields checks a Spend as sent (PRD §8.10, §6.9): the amount as a Salary's, one of the eight Categories, a
// date from earliestSpend to today in the Member's zone (A3), and a note of at most 80 NFC code points.
func (l member) spendFields(m *moneyrecapv1.Money, c moneyrecapv1.Category, day, note string) (*big.Rat, time.Time, string, error) {
	amount, err := l.amount(m)
	if err != nil {
		return nil, time.Time{}, "", err
	}
	if _, ok := moneyrecapv1.Category_name[int32(c)]; !ok || c == moneyrecapv1.Category_CATEGORY_UNSPECIFIED {
		return nil, time.Time{}, "", errors.New("category is not one of the eight")
	}
	date, err := time.Parse(time.DateOnly, day)
	if err != nil || date.After(l.today) || date.Before(l.earliestSpend()) {
		return nil, time.Time{}, "", errors.New("date is not YYYY-MM-DD in the last 60 days including today, from the sign-up month on")
	}
	note = norm.NFC.String(note)
	if utf8.RuneCountInString(note) > 80 {
		return nil, time.Time{}, "", errors.New("note is more than 80 characters")
	}
	return amount, date, note, nil
}

// earliestSpend is the first date a Spend may carry (PRD §8.10, D72): 59 days before today, or the 1st of the
// sign-up month when that is later.
func (l member) earliestSpend() time.Time {
	first := monthOf(l.signedUp).day(1)
	if back := l.today.AddDate(0, 0, -59); back.After(first) {
		return back
	}
	return first
}
