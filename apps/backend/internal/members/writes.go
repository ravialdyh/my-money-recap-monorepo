package members

import (
	"context"
	"errors"
	"time"
	_ "time/tzdata"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
)

// SetTimezone stores the Member's IANA zone (A3), checked against the zone data Go links in (time/tzdata;
// the image has none), and on the first call stamps signed_up_on, created_at's local date in that zone,
// which no later call changes (PRD §7).
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
		row, err := scanRow(tx.QueryRow(ctx, `UPDATE member SET timezone = $2, signed_up_on = coalesce(signed_up_on, $3::date) WHERE id = $1 RETURNING `+rowColumns,
			m.ID, req.Msg.Timezone, created.In(loc).Format(time.DateOnly)))
		answer.Profile = profile(row)
		return err
	})
	if errors.Is(err, idempotency.ErrReused) {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(answer), nil
}

// zone loads an IANA zone name. "" and "Local" load too but name no zone, so both are refused.
func zone(name string) (*time.Location, error) {
	loc, err := time.LoadLocation(name)
	if err != nil || name == "" || name == "Local" {
		return nil, errors.New("timezone is not an IANA zone name")
	}
	return loc, nil
}
