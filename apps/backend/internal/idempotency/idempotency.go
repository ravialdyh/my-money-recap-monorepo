package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
)

var ErrReused = errors.New("idempotency_key was already used for a different request")

func Write(ctx context.Context, db *pgxpool.Pool, memberID, key string, req, answer proto.Message, write func(pgx.Tx) error) error {
	sum, err := RequestHash(req)
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
				return ErrReused
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

func RequestHash(req proto.Message) ([]byte, error) {
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
