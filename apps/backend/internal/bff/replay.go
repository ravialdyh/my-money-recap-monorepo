package bff

import (
	"bytes"
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/idempotency"
)

// §4.4's idempotency fast path: an answer lives 24 hours, as the services' own keys do (A14), and
// replayTimeout bounds each Redis call as memberLimitTimeout does.
const (
	replayTTL     = 24 * time.Hour
	replayTimeout = 250 * time.Millisecond
)

// memberWrites are the Member writes the fast path answers, each with how its stored answer is read
// back. A write is covered by being listed; RequestMagicLink is not, having no Member (A6, A14).
var memberWrites = map[string]func([]byte) (connect.AnyResponse, error){
	moneyrecapv1connect.BffServiceSetTimezoneProcedure:    stored[moneyrecapv1.BffServiceSetTimezoneResponse],
	moneyrecapv1connect.BffServiceSaveProfileProcedure:    stored[moneyrecapv1.BffServiceSaveProfileResponse],
	moneyrecapv1connect.BffServiceSaveSalaryProcedure:     stored[moneyrecapv1.BffServiceSaveSalaryResponse],
	moneyrecapv1connect.BffServiceSaveTargetProcedure:     stored[moneyrecapv1.BffServiceSaveTargetResponse],
	moneyrecapv1connect.BffServiceCancelDeletionProcedure: stored[moneyrecapv1.BffServiceCancelDeletionResponse],
	moneyrecapv1connect.BffServiceLogSpendProcedure:       stored[moneyrecapv1.BffServiceLogSpendResponse],
	moneyrecapv1connect.BffServiceEditSpendProcedure:      stored[moneyrecapv1.BffServiceEditSpendResponse],
	moneyrecapv1connect.BffServiceDeleteSpendProcedure:    stored[moneyrecapv1.BffServiceDeleteSpendResponse],
	moneyrecapv1connect.BffServiceRegisterDeviceProcedure: stored[moneyrecapv1.BffServiceRegisterDeviceResponse],
	moneyrecapv1connect.BffServiceChangeCurrencyProcedure: stored[moneyrecapv1.BffServiceChangeCurrencyResponse],
}

// stored reads an answer's bytes back into the response type its procedure's handler returns.
func stored[T any, P interface {
	*T
	proto.Message
}](b []byte) (connect.AnyResponse, error) {
	msg := P(new(T))
	if err := proto.Unmarshal(b, msg); err != nil {
		return nil, err
	}
	return connect.NewResponse((*T)(msg)), nil
}

// Replay, placed after MemberLimit, answers a listed write from idem:{subject}:{key} when the stored
// request hash matches, and stores each successful answer. A different hash goes on to the owning
// service, which refuses it. Redis failing fails open (§4.4).
func Replay(cache redis.UniversalClient) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			read, listed := memberWrites[req.Spec().Procedure]
			msg, isMessage := req.Any().(proto.Message)
			subject, key := auth.From(ctx).Subject, req.Header().Get("Idempotency-Key")
			if !listed || !isMessage || subject == "" || key == "" {
				return next(ctx, req)
			}
			sum, err := idempotency.RequestHash(msg)
			if err != nil {
				logOptionalCallDegraded("redis.idempotency", "the write went to its service", err)
				return next(ctx, req)
			}
			name := "idem:" + subject + ":" + key
			if res, err := replayed(ctx, cache, name, sum, read); err != nil {
				logOptionalCallDegraded("redis.idempotency", "the write went to its service", err)
			} else if res != nil {
				return res, nil
			}
			res, err := next(ctx, req)
			if err != nil {
				return nil, err
			}
			if err := keep(ctx, cache, name, sum, res); err != nil {
				logOptionalCallDegraded("redis.idempotency", "the answer was not stored", err)
			}
			return res, nil
		}
	}
}

// replayed is the answer stored under name for the request hashed sum; nil with no error when none is.
func replayed(ctx context.Context, cache redis.UniversalClient, name string, sum []byte, read func([]byte) (connect.AnyResponse, error)) (connect.AnyResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, replayTimeout)
	defer cancel()
	b, err := cache.Get(ctx, name).Bytes()
	if errors.Is(err, redis.Nil) || (err == nil && !bytes.HasPrefix(b, sum)) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return read(b[len(sum):])
}

// keep stores the request hash followed by the answer's bytes under name for replayTTL.
func keep(ctx context.Context, cache redis.UniversalClient, name string, sum []byte, res connect.AnyResponse) error {
	msg, isMessage := res.Any().(proto.Message)
	if !isMessage {
		return errors.New("the answer is not a protobuf message")
	}
	answer, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, replayTimeout)
	defer cancel()
	return cache.Set(ctx, name, append(sum, answer...), replayTTL).Err()
}
