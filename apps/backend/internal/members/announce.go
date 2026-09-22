package members

import (
	"context"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
)

// facts is the replica other services keep of a Member (MemberFacts, ARCHITECTURE §7.1).
func facts(m Row) *moneyrecapv1.MemberFacts {
	return &moneyrecapv1.MemberFacts{Timezone: m.Timezone, SignedUpOn: m.SignedUpOn, CurrencyCode: m.CurrencyCode, DisplayName: m.DisplayName, Email: m.Email}
}

// locked reads the Member's row and holds it until the write's transaction ends, so what the write changed is
// judged against the row it actually replaced.
func locked(ctx context.Context, tx pgx.Tx, id string) (Row, error) {
	return scanRow(tx.QueryRow(ctx, `SELECT `+rowColumns+` FROM member WHERE id = $1 FOR UPDATE`, id))
}

// announce stages members.updated when a write changed any of the Member's facts; an Avatar-only save changes
// none and publishes nothing (spec #92).
func announce(ctx context.Context, tx pgx.Tx, before, after Row) error {
	if proto.Equal(facts(before), facts(after)) {
		return nil
	}
	return stage(ctx, tx, "members.updated", after)
}

// stage writes the event carrying m's facts to the outbox inside tx, with the change it describes (§4.3).
func stage(ctx context.Context, tx pgx.Tx, key string, m Row) error {
	return events.Stage(ctx, tx, m.ID, key, func(env *moneyrecapv1.EventEnvelope) proto.Message {
		if key == "members.created" {
			return &moneyrecapv1.MembersCreated{Envelope: env, Facts: facts(m)}
		}
		return &moneyrecapv1.MembersUpdated{Envelope: env, Facts: facts(m)}
	})
}
