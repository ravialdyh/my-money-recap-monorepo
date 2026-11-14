package members

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/answers"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
)

func (s *Service) SetTimezone(ctx context.Context, req *connect.Request[moneyrecapv1.SetTimezoneRequest]) (*connect.Response[moneyrecapv1.SetTimezoneResponse], error) {
	loc, err := zone(req.Msg.Timezone)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	m, err := s.own(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.SetTimezoneResponse{}
	err = idempotency.Write(ctx, s.DB, m.ID, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		var created time.Time
		if err := tx.QueryRow(ctx, `SELECT created_at FROM member WHERE id = $1`, m.ID).Scan(&created); err != nil {
			return err
		}
		before, err := locked(ctx, tx, m.ID)
		if err != nil {
			return err
		}
		after, err := scanRow(tx.QueryRow(ctx, `UPDATE member SET timezone = $2, signed_up_on = coalesce(signed_up_on, $3::date) WHERE id = $1 RETURNING `+rowColumns,
			m.ID, req.Msg.Timezone, created.In(loc).Format(time.DateOnly)))
		if err != nil {
			return err
		}
		answer.Profile = profile(after)
		return announce(ctx, tx, before, after)
	})
	if errors.Is(err, idempotency.ErrReused) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func zone(name string) (*time.Location, error) {
	loc, err := time.LoadLocation(name)
	if err != nil || name == "" || name == "Local" {
		return nil, errors.New("timezone is not an IANA zone name")
	}
	return loc, nil
}

func (s *Service) SaveProfile(ctx context.Context, req *connect.Request[moneyrecapv1.SaveProfileRequest]) (*connect.Response[moneyrecapv1.SaveProfileResponse], error) {
	msg := req.Msg
	if msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	name, err := answers.Check(msg.DisplayName, msg.AvatarKey, msg.CurrencyCode, msg.OnboardingStep)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	m, err := s.own(ctx, msg.MemberId)
	if err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.SaveProfileResponse{}
	err = idempotency.Write(ctx, s.DB, m.ID, msg.IdempotencyKey, msg, answer, func(tx pgx.Tx) error {
		before, err := locked(ctx, tx, m.ID)
		if err != nil {
			return err
		}
		after, err := scanRow(tx.QueryRow(ctx, `
			UPDATE member SET display_name = coalesce($2, display_name), avatar_key = coalesce($3, avatar_key),
				currency_code = coalesce($4, currency_code), onboarding_step = greatest(onboarding_step, $5)
			WHERE id = $1 RETURNING `+rowColumns, m.ID, name, msg.AvatarKey, msg.CurrencyCode, msg.OnboardingStep))
		if err != nil {
			return err
		}
		answer.Profile = profile(after)
		return announce(ctx, tx, before, after)
	})
	if errors.Is(err, idempotency.ErrReused) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

func (s *Service) RequestDeletion(ctx context.Context, req *connect.Request[moneyrecapv1.RequestDeletionRequest]) (*connect.Response[moneyrecapv1.RequestDeletionResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	m, err := s.own(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.RequestDeletionResponse{}
	err = idempotency.Write(ctx, s.DB, m.ID, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		before, err := locked(ctx, tx, m.ID)
		if err != nil {
			return err
		}
		answer.Profile = profile(before)
		if before.DeletionRequestedAt != nil {
			return nil
		}
		after, err := scanRow(tx.QueryRow(ctx, `UPDATE member SET deletion_requested_at = now() WHERE id = $1 RETURNING `+rowColumns, m.ID))
		if err != nil {
			return err
		}
		answer.Profile = profile(after)
		return events.Stage(ctx, tx, m.ID, "members.deletion.requested", func(env *moneyrecapv1.EventEnvelope) proto.Message {
			return &moneyrecapv1.MembersDeletionRequested{Envelope: env}
		})
	})
	if errors.Is(err, idempotency.ErrReused) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

const restoreWindow = 24 * time.Hour

var errWindowClosed = errors.New("the deletion request is over 24 hours old")

func (s *Service) CancelDeletion(ctx context.Context, req *connect.Request[moneyrecapv1.CancelDeletionRequest]) (*connect.Response[moneyrecapv1.CancelDeletionResponse], error) {
	if req.Msg.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	m, err := s.own(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.CancelDeletionResponse{}
	err = idempotency.Write(ctx, s.DB, m.ID, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		before, err := locked(ctx, tx, m.ID)
		if err != nil {
			return err
		}
		answer.Profile = profile(before)
		if before.DeletionRequestedAt == nil {
			return nil
		}
		if time.Since(*before.DeletionRequestedAt) >= restoreWindow {
			return errWindowClosed
		}
		after, err := scanRow(tx.QueryRow(ctx, `UPDATE member SET deletion_requested_at = NULL WHERE id = $1 RETURNING `+rowColumns, m.ID))
		if err != nil {
			return err
		}
		answer.Profile = profile(after)
		return events.Stage(ctx, tx, m.ID, "members.deletion.cancelled", func(env *moneyrecapv1.EventEnvelope) proto.Message {
			return &moneyrecapv1.MembersDeletionCancelled{Envelope: env}
		})
	})
	switch {
	case errors.Is(err, idempotency.ErrReused):
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, errWindowClosed):
		return nil, connect.NewError(connect.CodeFailedPrecondition, err)
	case err != nil:
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}
