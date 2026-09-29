package notify

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

var hexToken = regexp.MustCompile(`^[0-9a-f]+$`)

// RegisterDevice upserts the phone's row on its APNs token under the Member's key (A14): a token another Member held
// moves to this one, and a Member may hold several (ARCHITECTURE §7.3).
func (s *Service) RegisterDevice(ctx context.Context, req *connect.Request[moneyrecapv1.RegisterDeviceRequest]) (*connect.Response[moneyrecapv1.RegisterDeviceResponse], error) {
	m := req.Msg
	if m.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	token := strings.ToLower(m.ApnsToken)
	if !hexToken.MatchString(token) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("apns_token is not hex"))
	}
	if err := s.owns(ctx, m.MemberId); err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.RegisterDeviceResponse{}
	err := s.write(ctx, m.MemberId, m.IdempotencyKey, m, answer, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO device (apns_token, member_id) VALUES ($1, $2)
			ON CONFLICT (apns_token) DO UPDATE SET member_id = EXCLUDED.member_id, registered_at = now()`, token, m.MemberId)
		return err
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(answer), nil
}
