package bff

import (
	"context"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// §4.4's per-Member limit: 120 requests per subject per fixed minute (#97), the key living two
// minutes. memberLimitTimeout bounds the count so a Redis that hangs costs a request a quarter of a
// second, not the one-second budget of the calls behind it (§6.1).
const (
	memberLimitPerMinute = 120
	memberLimitTTL       = 2 * time.Minute
	memberLimitTimeout   = 250 * time.Millisecond
)

// MemberLimit, placed after auth's interceptor, refuses a subject's request past the ceiling with
// resource_exhausted. It fails open, so a cache that is down never locks a Member out (§4.4, A6). No
// subject means RequestMagicLink, which has its own limit.
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

// countMember is INCR and EXPIRE on rl:{subject}:{minute}, the minute counted from the Unix epoch.
func countMember(ctx context.Context, limit Limiter, subject string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, memberLimitTimeout)
	defer cancel()
	return limit.Count(ctx, fmt.Sprintf("rl:%s:%d", subject, time.Now().Unix()/60), memberLimitTTL)
}
