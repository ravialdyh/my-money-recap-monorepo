package members

import (
	"context"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
)

func facts(m Row) *moneyrecapv1.MemberFacts {
	return &moneyrecapv1.MemberFacts{Timezone: m.Timezone, SignedUpOn: m.SignedUpOn, CurrencyCode: m.CurrencyCode, DisplayName: m.DisplayName, Email: m.Email}
}

func locked(ctx context.Context, tx pgx.Tx, id string) (Row, error) {
	return scanRow(tx.QueryRow(ctx, `SELECT `+rowColumns+` FROM member WHERE id = $1 FOR UPDATE`, id))
}

func announce(ctx context.Context, tx pgx.Tx, before, after Row) error {
	if proto.Equal(facts(before), facts(after)) {
		return nil
	}
	return stage(ctx, tx, "members.updated", after)
}

func stage(ctx context.Context, tx pgx.Tx, key string, m Row) error {
	return events.Stage(ctx, tx, m.ID, key, func(env *moneyrecapv1.EventEnvelope) proto.Message {
		if key == "members.created" {
			return &moneyrecapv1.MembersCreated{Envelope: env, Facts: facts(m)}
		}
		return &moneyrecapv1.MembersUpdated{Envelope: env, Facts: facts(m)}
	})
}
