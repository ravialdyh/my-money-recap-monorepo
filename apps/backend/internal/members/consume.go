package members

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/answers"
)

// Queue is members' queue of ledger's events, and Keys what it binds (ARCHITECTURE §4.3, §7.1 "Consumes").
const Queue = "members.ledger"

var Keys = []string{"ledger.currency.converted"}

// ApplyConversion sets the Currency ledger converted the Member's amounts to and stages members.updated, unless a
// newer conversion already set one (currency_at), so a late delivery never takes the code back (PRD §8.9).
func ApplyConversion(ctx context.Context, tx pgx.Tx, d amqp.Delivery) error {
	var e moneyrecapv1.LedgerCurrencyConverted
	if err := proto.Unmarshal(d.Body, &e); err != nil {
		return err
	}
	env := e.GetEnvelope()
	if env.GetMemberId() == "" || env.GetOccurredAt() == nil || answers.CurrencyCode(e.ToCurrencyCode) != nil {
		return errors.New("the event names no Member, no time or no Currency")
	}
	before, err := locked(ctx, tx, env.MemberId)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	after, err := scanRow(tx.QueryRow(ctx, `UPDATE member SET currency_code = $2, currency_at = $3
		WHERE id = $1 AND (currency_at IS NULL OR currency_at < $3) RETURNING `+rowColumns, env.MemberId, e.ToCurrencyCode, env.OccurredAt.AsTime()))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return announce(ctx, tx, before, after)
}
