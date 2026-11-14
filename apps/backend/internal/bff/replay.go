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

const (
	replayTTL     = 24 * time.Hour
	replayTimeout = 250 * time.Millisecond
)

var memberWrites = map[string]func([]byte) (connect.AnyResponse, error){
	moneyrecapv1connect.BffServiceSetTimezoneProcedure:           stored[moneyrecapv1.BffServiceSetTimezoneResponse],
	moneyrecapv1connect.BffServiceSaveProfileProcedure:           stored[moneyrecapv1.BffServiceSaveProfileResponse],
	moneyrecapv1connect.BffServiceSaveSalaryProcedure:            stored[moneyrecapv1.BffServiceSaveSalaryResponse],
	moneyrecapv1connect.BffServiceSaveTargetProcedure:            stored[moneyrecapv1.BffServiceSaveTargetResponse],
	moneyrecapv1connect.BffServiceCancelDeletionProcedure:        stored[moneyrecapv1.BffServiceCancelDeletionResponse],
	moneyrecapv1connect.BffServiceLogSpendProcedure:              stored[moneyrecapv1.BffServiceLogSpendResponse],
	moneyrecapv1connect.BffServiceEditSpendProcedure:             stored[moneyrecapv1.BffServiceEditSpendResponse],
	moneyrecapv1connect.BffServiceDeleteSpendProcedure:           stored[moneyrecapv1.BffServiceDeleteSpendResponse],
	moneyrecapv1connect.BffServiceRegisterDeviceProcedure:        stored[moneyrecapv1.BffServiceRegisterDeviceResponse],
	moneyrecapv1connect.BffServiceChangeCurrencyProcedure:        stored[moneyrecapv1.BffServiceChangeCurrencyResponse],
	moneyrecapv1connect.BffServiceSaveNotificationPrefsProcedure: stored[moneyrecapv1.BffServiceSaveNotificationPrefsResponse],
	moneyrecapv1connect.BffServiceRequestExportProcedure:         stored[moneyrecapv1.BffServiceRequestExportResponse],
	moneyrecapv1connect.BffServiceForgetDeviceProcedure:          stored[moneyrecapv1.BffServiceForgetDeviceResponse],
	moneyrecapv1connect.BffServiceRequestDeletionProcedure:       stored[moneyrecapv1.BffServiceRequestDeletionResponse],
}

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
