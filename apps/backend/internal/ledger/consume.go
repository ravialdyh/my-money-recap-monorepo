package ledger

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// Queue is ledger's queue of members' events, and Keys what it binds (ARCHITECTURE §4.3, §7.2 "Consumes").
const Queue = "ledger.members"

var Keys = []string{"members.created", "members.updated"}

// ApplyMemberFacts keeps the replica from members.created and members.updated (one shape): an older event, or a
// Currency older than RelabelCurrency's, is skipped, an empty fact never clears one, and the subject stays (§4.2).
func ApplyMemberFacts(ctx context.Context, tx pgx.Tx, d amqp.Delivery) error {
	var e moneyrecapv1.MembersUpdated
	if err := proto.Unmarshal(d.Body, &e); err != nil {
		return err
	}
	env, f := e.GetEnvelope(), e.GetFacts()
	if env.GetMemberId() == "" || env.GetOccurredAt() == nil || f == nil {
		return errors.New("the event names no Member, no time or no facts")
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO member_facts (member_id, timezone, signed_up_on, currency_code, facts_at)
		VALUES ($1, nullif($2, ''), nullif($3, '')::date, nullif($4, ''), $5)
		ON CONFLICT (member_id) DO UPDATE SET
			timezone = coalesce(EXCLUDED.timezone, member_facts.timezone),
			signed_up_on = coalesce(EXCLUDED.signed_up_on, member_facts.signed_up_on),
			currency_code = CASE WHEN member_facts.currency_at > EXCLUDED.facts_at THEN member_facts.currency_code
				ELSE coalesce(EXCLUDED.currency_code, member_facts.currency_code) END,
			facts_at = EXCLUDED.facts_at
		WHERE member_facts.facts_at IS NULL OR member_facts.facts_at <= EXCLUDED.facts_at`,
		env.MemberId, f.Timezone, f.SignedUpOn, f.CurrencyCode, env.OccurredAt.AsTime())
	return err
}
