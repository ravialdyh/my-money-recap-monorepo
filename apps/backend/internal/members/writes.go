package members

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"time"
	_ "time/tzdata"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// errKeyReused is a key sent again with a different request: answering the first call's result would
// silently drop the second write.
var errKeyReused = errors.New("idempotency_key was already used for a different request")

// keyedWrite runs write once per Member and key (A14): the key is claimed in the write's transaction, a
// repeat replays the stored answer and writes nothing, a key reused for another request is refused, and
// keys over 24 hours old (§4.4) are swept first.
func keyedWrite(ctx context.Context, db *pgxpool.Pool, memberID, key string, req, answer proto.Message, write func(pgx.Tx) error) error {
	sum, err := requestHash(req)
	if err != nil {
		return err
	}
	return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM idempotency WHERE created_at < now() - interval '24 hours'`); err != nil {
			return err
		}
		claimed, err := tx.Exec(ctx, `INSERT INTO idempotency (member_id, idempotency_key, request_hash) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, memberID, key, sum)
		if err != nil {
			return err
		}
		if claimed.RowsAffected() == 0 {
			var storedSum, stored []byte
			if err := tx.QueryRow(ctx, `SELECT request_hash, response FROM idempotency WHERE member_id = $1 AND idempotency_key = $2`, memberID, key).Scan(&storedSum, &stored); err != nil {
				return err
			}
			if !bytes.Equal(storedSum, sum) {
				return errKeyReused
			}
			return proto.Unmarshal(stored, answer)
		}
		if err := write(tx); err != nil {
			return err
		}
		stored, err := proto.Marshal(answer)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE idempotency SET response = $3 WHERE member_id = $1 AND idempotency_key = $2`, memberID, key, stored)
		return err
	})
}

// requestHash identifies a request by its message name and deterministic bytes, so the same key on two
// different calls, or on one call with other values, cannot pass for a retry.
func requestHash(req proto.Message) ([]byte, error) {
	b, err := proto.MarshalOptions{Deterministic: true}.Marshal(req)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	h.Write([]byte(proto.MessageName(req)))
	h.Write([]byte{0})
	h.Write(b)
	return h.Sum(nil), nil
}

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
	err = keyedWrite(ctx, s.DB, m.ID, req.Msg.IdempotencyKey, req.Msg, answer, func(tx pgx.Tx) error {
		var created time.Time
		if err := tx.QueryRow(ctx, `SELECT created_at FROM member WHERE id = $1`, m.ID).Scan(&created); err != nil {
			return err
		}
		row, err := scanRow(tx.QueryRow(ctx, `UPDATE member SET timezone = $2, signed_up_on = coalesce(signed_up_on, $3::date) WHERE id = $1 RETURNING `+rowColumns,
			m.ID, req.Msg.Timezone, created.In(loc).Format(time.DateOnly)))
		answer.Profile = profile(row)
		return err
	})
	if errors.Is(err, errKeyReused) {
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
