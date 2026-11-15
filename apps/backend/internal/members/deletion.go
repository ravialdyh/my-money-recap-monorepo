package members

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
)

const PurgesQueue = "members.purges"

var PurgesKeys = []string{"ledger.member.purged", "notify.member.purged"}

const (
	sweepEvery = 10 * time.Minute
	stallAfter = time.Hour
)

const tokenLifetime = 5 * time.Minute

var purgedColumn = map[string]string{"ledger.member.purged": "ledger_purged_at", "notify.member.purged": "notify_purged_at"}

func (s *Service) Sweep(ctx context.Context) {
	every := time.NewTicker(sweepEvery)
	defer every.Stop()
	for {
		if err := s.DeleteDue(ctx, time.Now()); err != nil && ctx.Err() == nil {
			log.Printf("members: the deletion job: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-every.C:
		}
	}
}

func (s *Service) DeleteDue(ctx context.Context, now time.Time) error {
	rows, err := s.DB.Query(ctx, `SELECT id::text FROM member WHERE deletion_requested_at <= $1 AND deleted_at IS NULL`, now.Add(-restoreWindow))
	if err != nil {
		return err
	}
	due, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}
	for _, id := range due {
		if err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error { return tombstone(ctx, tx, id, now) }); err != nil {
			return err
		}
	}
	rows, err = s.DB.Query(ctx, `SELECT member_id::text, keycloak_subject FROM deletion_progress WHERE keycloak_deleted_at IS NULL`)
	if err != nil {
		return err
	}
	held, err := pgx.CollectRows(rows, pgx.RowToStructByPos[struct{ ID, Subject string }])
	if err != nil {
		return err
	}
	for _, m := range held {
		if err := s.Keycloak.DeleteAccount(ctx, m.Subject); err != nil {
			log.Printf("members: deleting the Keycloak account of %s: %v", m.ID, err)
			continue
		}
		if err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error { return stamp(ctx, tx, m.ID, "keycloak_deleted_at") }); err != nil {
			return err
		}
	}
	if err := s.finish(ctx, now); err != nil {
		return err
	}
	rows, err = s.DB.Query(ctx, `SELECT member_id::text FROM deletion_progress WHERE deleted_at <= $1`, now.Add(-stallAfter))
	if err != nil {
		return err
	}
	stalled, err := pgx.CollectRows(rows, pgx.RowTo[string])
	for _, id := range stalled {
		log.Printf("members: the deletion of %s is incomplete an hour after its tombstone", id)
	}
	return err
}

func tombstone(ctx context.Context, tx pgx.Tx, id string, now time.Time) error {
	tag, err := tx.Exec(ctx, `WITH gone AS (UPDATE member SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL RETURNING id, keycloak_subject, deleted_at)
		INSERT INTO deletion_progress (member_id, keycloak_subject, deleted_at) SELECT * FROM gone`, id, now)
	if err != nil || tag.RowsAffected() == 0 {
		return err
	}
	return events.Stage(ctx, tx, id, "members.deleted", func(env *moneyrecapv1.EventEnvelope) proto.Message {
		return &moneyrecapv1.MembersDeleted{Envelope: env}
	})
}

func ApplyPurged(ctx context.Context, tx pgx.Tx, d amqp.Delivery) error {
	var e moneyrecapv1.LedgerMemberPurged
	if err := proto.Unmarshal(d.Body, &e); err != nil {
		return err
	}
	column, id := purgedColumn[d.RoutingKey], e.GetEnvelope().GetMemberId()
	if column == "" || id == "" {
		return errors.New("the event names no Member or is not a purged event")
	}
	return stamp(ctx, tx, id, column)
}

func stamp(ctx context.Context, tx pgx.Tx, id, column string) error {
	_, err := tx.Exec(ctx, `UPDATE deletion_progress SET `+column+` = coalesce(`+column+`, now()) WHERE member_id = $1`, id)
	return err
}

func (s *Service) finish(ctx context.Context, now time.Time) error {
	_, err := s.DB.Exec(ctx, `WITH done AS (
			DELETE FROM deletion_progress WHERE keycloak_deleted_at <= $1
				AND ledger_purged_at IS NOT NULL AND notify_purged_at IS NOT NULL
			RETURNING member_id),
		keys AS (DELETE FROM idempotency WHERE member_id IN (SELECT member_id FROM done)),
		staged AS (DELETE FROM outbox WHERE member_id IN (SELECT member_id FROM done))
		DELETE FROM member WHERE id IN (SELECT member_id FROM done)`, now.Add(-tokenLifetime))
	return err
}
