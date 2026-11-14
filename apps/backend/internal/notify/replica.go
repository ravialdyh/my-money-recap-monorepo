package notify

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// MembersQueue is notify's queue of members' events, and MembersKeys what it binds (ARCHITECTURE §4.3, §7.3
// "Consumes"): the facts it copies and the deletion-pending flag.
const MembersQueue = "notify.members"

var MembersKeys = []string{"members.created", "members.updated", "members.deletion.requested", "members.deletion.cancelled"}

// ApplyMemberEvent keeps the replica from MembersKeys: the zone and the email from the facts, and the
// deletion-pending flag. Each has its own stamp, so an older event never overrides a newer one of its kind.
func ApplyMemberEvent(ctx context.Context, tx pgx.Tx, d amqp.Delivery) error {
	switch d.RoutingKey {
	case "members.created", "members.updated":
		return applyFacts(ctx, tx, d.Body)
	case "members.deletion.requested":
		return flagDeletion(ctx, tx, d.Body, true)
	case "members.deletion.cancelled":
		return flagDeletion(ctx, tx, d.Body, false)
	}
	return fmt.Errorf("notify.members does not bind %q", d.RoutingKey)
}

// applyFacts copies the zone and the email of members.created and members.updated (one shape); an empty fact
// never clears one, the Display Name is not kept, since no push prints it, and a new zone moves both next-fire times.
func applyFacts(ctx context.Context, tx pgx.Tx, body []byte) error {
	var e moneyrecapv1.MembersUpdated
	if err := proto.Unmarshal(body, &e); err != nil {
		return err
	}
	env, f := e.GetEnvelope(), e.GetFacts()
	if env.GetMemberId() == "" || env.GetOccurredAt() == nil || f == nil {
		return errors.New("the event names no Member, no time or no facts")
	}
	var moved bool
	err := tx.QueryRow(ctx, `
		WITH old AS (SELECT timezone FROM member_facts WHERE member_id = $1)
		INSERT INTO member_facts (member_id, timezone, email, facts_at) VALUES ($1, nullif($2, ''), nullif($3, ''), $4)
		ON CONFLICT (member_id) DO UPDATE SET
			timezone = coalesce(EXCLUDED.timezone, member_facts.timezone),
			email = coalesce(EXCLUDED.email, member_facts.email),
			facts_at = EXCLUDED.facts_at
		WHERE member_facts.facts_at IS NULL OR member_facts.facts_at <= EXCLUDED.facts_at
		RETURNING member_facts.timezone IS DISTINCT FROM (SELECT timezone FROM old)`,
		env.MemberId, f.Timezone, f.Email, env.OccurredAt.AsTime()).Scan(&moved)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !moved) {
		return nil
	}
	if err != nil {
		return err
	}
	return reschedule(ctx, tx, env.MemberId, time.Now())
}

// flagDeletion sets or clears the deletion-pending flag from members.deletion.requested or
// members.deletion.cancelled (one shape), unless a newer one of the two already has.
func flagDeletion(ctx context.Context, tx pgx.Tx, body []byte, pending bool) error {
	var e moneyrecapv1.MembersDeletionRequested
	if err := proto.Unmarshal(body, &e); err != nil {
		return err
	}
	env := e.GetEnvelope()
	if env.GetMemberId() == "" || env.GetOccurredAt() == nil {
		return errors.New("the event names no Member or no time")
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO member_facts (member_id, deletion_pending, deletion_at) VALUES ($1, $2, $3)
		ON CONFLICT (member_id) DO UPDATE SET deletion_pending = EXCLUDED.deletion_pending, deletion_at = EXCLUDED.deletion_at
		WHERE member_facts.deletion_at IS NULL OR member_facts.deletion_at <= EXCLUDED.deletion_at`,
		env.MemberId, pending, env.OccurredAt.AsTime())
	return err
}
