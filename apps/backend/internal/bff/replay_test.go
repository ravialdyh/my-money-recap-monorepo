package bff

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// stackRedis is the stack's Redis at BFF_TEST_REDIS; the test skips without it.
func stackRedis(t *testing.T) redis.UniversalClient {
	t.Helper()
	addr := os.Getenv("BFF_TEST_REDIS")
	if addr == "" {
		t.Skip("BFF_TEST_REDIS not set")
	}
	cache := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = cache.Close() })
	return cache
}

// replayClient is the BFF behind Replay alone, as a fresh subject each time, so no stored answer leaks
// between tests.
func replayClient(t *testing.T, members *fakeMembers, cache redis.UniversalClient) moneyrecapv1connect.BffServiceClient {
	t.Helper()
	subject := "sub-" + rand.Text()
	signedIn := connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			return next(auth.NewContext(ctx, auth.Claims{Subject: subject, Raw: "tok"}), req)
		}
	})
	mux := http.NewServeMux()
	mux.Handle(moneyrecapv1connect.NewBffServiceHandler(newService(t, members, StubNotify{}), connect.WithInterceptors(signedIn, Replay(cache))))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return moneyrecapv1connect.NewBffServiceClient(srv.Client(), srv.URL)
}

// setZone sends SetTimezone under key and says whether members was reached at all.
func setZone(t *testing.T, client moneyrecapv1connect.BffServiceClient, members *fakeMembers, key, zone string) (*moneyrecapv1.Profile, bool, error) {
	t.Helper()
	members.gotToken, members.gotZone = "", nil
	res, err := client.SetTimezone(context.Background(), keyed(&moneyrecapv1.BffServiceSetTimezoneRequest{Timezone: zone}, key))
	reached := members.gotToken != "" || members.gotZone != nil
	if err != nil {
		return nil, reached, err
	}
	return res.Msg.Profile, reached, nil
}

func TestReplayAnswersARetryFromRedisAndSendsOtherValuesOn(t *testing.T) {
	members := &fakeMembers{}
	client := replayClient(t, members, stackRedis(t))
	key := rand.Text()
	first, reached, err := setZone(t, client, members, key, "Asia/Jakarta")
	if err != nil || !reached {
		t.Fatalf("first call: %v, reached %v", err, reached)
	}
	again, reached, err := setZone(t, client, members, key, "Asia/Jakarta")
	if err != nil || reached || !proto.Equal(first, again) {
		t.Fatalf("the retry: %v, %v, reached %v; want the first answer without members", again, err, reached)
	}
	if _, reached, _ := setZone(t, client, members, key, "UTC"); !reached {
		t.Error("the key reused for another zone was answered from Redis, want it sent to members")
	}
	if _, _, err := setZone(t, client, members, "", "UTC"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("a keyless call: %v, want the handler's invalid_argument", err)
	}
}

func TestReplayNeverStoresAFailure(t *testing.T) {
	members := &fakeMembers{createErr: connect.NewError(connect.CodeUnavailable, errors.New("down"))}
	client := replayClient(t, members, stackRedis(t))
	key := rand.Text()
	if _, _, err := setZone(t, client, members, key, "UTC"); connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("members down: %v, want unavailable", err)
	}
	members.createErr = nil
	if _, reached, err := setZone(t, client, members, key, "UTC"); err != nil || !reached {
		t.Fatalf("the retry after a failure: %v, reached %v; want members to answer it", err, reached)
	}
	if _, reached, err := setZone(t, client, members, key, "UTC"); err != nil || reached {
		t.Fatalf("the retry after a success: %v, reached %v; want Redis to answer it", err, reached)
	}
}

// A Redis that answers nothing sends every write to its service, twice under one key, and costs each
// Redis call at most replayTimeout (§4.4).
func TestReplayFailsOpenWhenRedisIsUnreachable(t *testing.T) {
	unreachable := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1})
	t.Cleanup(func() { _ = unreachable.Close() })
	members := &fakeMembers{}
	client := replayClient(t, members, unreachable)
	key := rand.Text()
	for i := range 2 {
		start := time.Now()
		if _, reached, err := setZone(t, client, members, key, "UTC"); err != nil || !reached {
			t.Fatalf("call %d: %v, reached %v; want members to answer", i+1, err, reached)
		}
		if took := time.Since(start); took > 2*replayTimeout+200*time.Millisecond {
			t.Errorf("call %d took %s", i+1, took)
		}
	}
}

// Each listed write reads its answer back as the type its handler returns, or Connect would refuse the
// replay; RequestMagicLink is not listed.
func TestEveryListedWriteReadsBackItsOwnResponseType(t *testing.T) {
	methods := moneyrecapv1.File_money_recap_v1_bff_proto.Services().ByName("BffService").Methods()
	for procedure, read := range memberWrites {
		method := methods.ByName(protoreflect.Name(path.Base(procedure)))
		if method == nil {
			t.Fatalf("%s is not a BffService procedure", procedure)
		}
		res, err := read(nil)
		msg, isMessage := res.Any().(proto.Message)
		if err != nil || !isMessage || msg.ProtoReflect().Descriptor().FullName() != method.Output().FullName() {
			t.Errorf("%s reads back %T (%v), want %s", procedure, res.Any(), err, method.Output().FullName())
		}
	}
	if _, listed := memberWrites[moneyrecapv1connect.BffServiceRequestMagicLinkProcedure]; listed {
		t.Error("RequestMagicLink is listed")
	}
}
