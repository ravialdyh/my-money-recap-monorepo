package bff

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

const (
	memberLimitPerMinute = 120
	memberLimitTTL       = 2 * time.Minute
	memberLimitTimeout   = 250 * time.Millisecond
)

func MemberLimit(limit Limiter) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			subject := auth.From(ctx).Subject
			if subject == "" {
				return next(ctx, req)
			}
			count, err := countMember(ctx, limit, subject)
			if err != nil {
				logOptionalCallDegraded("redis.member-limit", "the request passed uncounted", err)
				return next(ctx, req)
			}
			if count > memberLimitPerMinute {
				return nil, connect.NewError(connect.CodeResourceExhausted, errors.New("too many requests this minute"))
			}
			return next(ctx, req)
		}
	}
}

func countMember(ctx context.Context, limit Limiter, subject string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, memberLimitTimeout)
	defer cancel()
	return limit.Count(ctx, fmt.Sprintf("rl:%s:%d", subject, time.Now().Unix()/60), memberLimitTTL)
}
