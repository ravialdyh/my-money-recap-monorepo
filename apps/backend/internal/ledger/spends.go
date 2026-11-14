package ledger

import (
	"context"
	"errors"
	"math/big"
	"slices"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"golang.org/x/text/unicode/norm"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
)

func (s *Service) LogSpend(ctx context.Context, req *connect.Request[moneyrecapv1.LogSpendRequest]) (*connect.Response[moneyrecapv1.LogSpendResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	amount, date, note, err := l.spendFields(req.Msg.Amount, req.Msg.Category, req.Msg.Date, req.Msg.Note)
	if err == nil && !l.datable(date) {
		err = errDate
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	answer := &moneyrecapv1.LogSpendResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		if err := l.relock(ctx, tx); err != nil {
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
		answer.CrossedLine = l.crossed(m, before, answer.Month)
		return l.announce(ctx, tx, "ledger.spend.logged", spend, answer.Month, answer.CrossedLine)
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func (s *Service) EditSpend(ctx context.Context, req *connect.Request[moneyrecapv1.EditSpendRequest]) (*connect.Response[moneyrecapv1.EditSpendResponse], error) {
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
	answer := &moneyrecapv1.EditSpendResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		i, err := l.lockSpend(ctx, tx, req.Msg.SpendId)
		if err != nil {
			return err
		}
		old := l.spends[i]
		if !date.Equal(old.date) && (!l.datable(date) || !l.datable(old.date)) {
			return connect.NewError(connect.CodeInvalidArgument, errDate)
		}
		m := monthOf(date)
		before := l.summary(m, l.today)
		spend := &moneyrecapv1.Spend{
			Id: old.spend.Id, Amount: l.cash.money(amount), OriginalAmount: l.cash.money(amount), Category: req.Msg.Category,
			Date: date.Format(time.DateOnly), Note: note, CreatedAt: old.spend.CreatedAt,
		}
		var updated time.Time
		if err := tx.QueryRow(ctx, `UPDATE spend SET amount = $3, original_amount = $3, original_currency = $4, category = $5, spent_on = $6, note = $7, updated_at = now()
			WHERE id = $1 AND member_id = $2 RETURNING updated_at`, spend.Id, l.id, spend.Amount.Amount, l.cash.code, int16(spend.Category), date, note).Scan(&updated); err != nil {
			return err
		}
		spend.UpdatedAt = timestamppb.New(updated)
		l.spends[i] = spendRow{amount: amount, date: date, spend: spend}
		answer.Spend, answer.Month = spend, l.summary(m, l.today)
		answer.CrossedLine = l.crossed(m, before, answer.Month)
		return l.announce(ctx, tx, "ledger.spend.edited", spend, answer.Month, answer.CrossedLine)
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func (s *Service) DeleteSpend(ctx context.Context, req *connect.Request[moneyrecapv1.DeleteSpendRequest]) (*connect.Response[moneyrecapv1.DeleteSpendResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.DeleteSpendResponse{}
	err = idempotency.Write(ctx, s.DB, l.id, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		i, err := l.lockSpend(ctx, tx, req.Msg.SpendId)
		if err != nil {
			return err
		}
		gone := l.spends[i]
		if _, err := tx.Exec(ctx, `UPDATE spend SET deleted_at = now() WHERE id = $1 AND member_id = $2`, gone.spend.Id, l.id); err != nil {
			return err
		}
		l.spends = slices.Delete(l.spends, i, i+1)
		answer.Spend, answer.Month = gone.spend, l.summary(monthOf(gone.date), l.today)
		return l.announce(ctx, tx, "ledger.spend.deleted", gone.spend, answer.Month, false)
	})
	if err != nil {
		return nil, writeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func (l member) announce(ctx context.Context, tx pgx.Tx, key string, spend *moneyrecapv1.Spend, month *moneyrecapv1.MonthSummary, crossed bool) error {
	return events.Stage(ctx, tx, l.id, key, func(env *moneyrecapv1.EventEnvelope) proto.Message {
		switch key {
		case "ledger.spend.logged":
			return &moneyrecapv1.LedgerSpendLogged{Envelope: env, Spend: spend, Month: month, CrossedLine: crossed}
		case "ledger.spend.edited":
			return &moneyrecapv1.LedgerSpendEdited{Envelope: env, Spend: spend, Month: month, CrossedLine: crossed}
		}
		return &moneyrecapv1.LedgerSpendDeleted{Envelope: env, Spend: spend, Month: month}
	})
}

func (l member) crossed(m month, before, after *moneyrecapv1.MonthSummary) bool {
	return m == monthOf(l.today) && before.TargetWarning == nil && after.TargetWarning != nil
}

func (l *member) lockSpend(ctx context.Context, tx pgx.Tx, id string) (int, error) {
	if err := l.relock(ctx, tx); err != nil {
		return 0, err
	}
	i := slices.IndexFunc(l.spends, func(r spendRow) bool { return r.spend.Id == id })
	if i < 0 {
		return 0, connect.NewError(connect.CodeNotFound, errors.New("no such spend"))
	}
	return i, nil
}

var errDate = errors.New("date is not YYYY-MM-DD in the last 60 days including today, from the sign-up month on")

func (l *member) relock(ctx context.Context, tx pgx.Tx) (err error) {
	if _, err = tx.Exec(ctx, `SELECT 1 FROM member_facts WHERE member_id = $1 FOR UPDATE`, l.id); err != nil {
		return err
	}
	l.spends, err = spends(ctx, tx, l.id, l.cash)
	return err
}

func (l member) spendFields(m *moneyrecapv1.Money, c moneyrecapv1.Category, day, note string) (*big.Rat, time.Time, string, error) {
	amount, err := l.amount(m)
	if err != nil {
		return nil, time.Time{}, "", err
	}
	if _, ok := moneyrecapv1.Category_name[int32(c)]; !ok || c == moneyrecapv1.Category_CATEGORY_UNSPECIFIED {
		return nil, time.Time{}, "", errors.New("category is not one of the eight")
	}
	date, err := time.Parse(time.DateOnly, day)
	if err != nil {
		return nil, time.Time{}, "", errDate
	}
	note = norm.NFC.String(note)
	if utf8.RuneCountInString(note) > 80 {
		return nil, time.Time{}, "", errors.New("note is more than 80 characters")
	}
	return amount, date, note, nil
}

func (l member) datable(d time.Time) bool { return !d.After(l.today) && !d.Before(l.earliestSpend()) }

func (l member) earliestSpend() time.Time {
	first := monthOf(l.signedUp).day(1)
	if back := l.today.AddDate(0, 0, -59); back.After(first) {
		return back
	}
	return first
}
