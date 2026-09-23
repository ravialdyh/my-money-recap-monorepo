package bff

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/redis/go-redis/v9"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// served is the procedure behind the limit: it counts the requests that reached it.
type served struct{ calls int }

func (s *served) call(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
	s.calls++
	return connect.NewResponse(&moneyrecapv1.GetHomeResponse{}), nil
}

func asMember(subject string) context.Context {
	return auth.NewContext(context.Background(), auth.Claims{Subject: subject, Raw: "tok"})
}

func TestMemberLimitRefusesTheRequestPastTheCeilingOfAMinute(t *testing.T) {
	limit, backend := &fakeLimit{}, &served{}
	handle := MemberLimit(limit)(backend.call)
	req := connect.NewRequest(&moneyrecapv1.GetHomeRequest{})
	for i := range memberLimitPerMinute {
		if _, err := handle(asMember("sub-1"), req); err != nil {
			t.Fatalf("request %d of the minute: %v", i+1, err)
		}
	}
	_, err := handle(asMember("sub-1"), req)
	if connect.CodeOf(err) != connect.CodeResourceExhausted {
		t.Fatalf("request %d: %v, want resource_exhausted", memberLimitPerMinute+1, err)
	}
	if _, err := handle(asMember("sub-2"), req); err != nil {
		t.Fatalf("another Member in the same minute: %v, want served", err)
	}
	if backend.calls != memberLimitPerMinute+1 {
		t.Fatalf("the procedure ran %d times, want %d", backend.calls, memberLimitPerMinute+1)
	}
	for key := range limit.counts {
		if !strings.HasPrefix(key, "rl:sub-") || strings.Count(key, ":") != 2 {
			t.Errorf("key %q, want rl:{subject}:{minute}", key)
		}
	}
}

// The open door: a Redis that answers nothing lets the same request through, where the magic-link
// limit refuses it (§4.4).
func TestMemberLimitFailsOpenWhenRedisIsUnreachable(t *testing.T) {
	backend := &served{}
	unreachable := RedisLimiter{Client: redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1})}
	t.Cleanup(func() { _ = unreachable.Client.Close() })
	start := time.Now()
	_, err := MemberLimit(unreachable)(backend.call)(asMember("sub-1"), connect.NewRequest(&moneyrecapv1.GetHomeRequest{}))
	if err != nil || backend.calls != 1 {
		t.Fatalf("err %v, %d calls; want the request served", err, backend.calls)
	}
	if took := time.Since(start); took > memberLimitTimeout+100*time.Millisecond {
		t.Errorf("an unreachable Redis cost the request %s", took)
	}
}

// refusing fails the test if anything is counted.
type refusing struct{ t *testing.T }

func (r refusing) Count(context.Context, string, time.Duration) (int64, error) {
	r.t.Error("a request without a subject was counted")
	return 0, fmt.Errorf("refused")
}

func TestMemberLimitLeavesTheMagicLinkRequestToItsOwnLimit(t *testing.T) {
	backend := &served{}
	req := connect.NewRequest(&moneyrecapv1.BffServiceRequestMagicLinkRequest{Email: "test@example.com"})
	if _, err := MemberLimit(refusing{t})(backend.call)(context.Background(), req); err != nil || backend.calls != 1 {
		t.Fatalf("err %v, %d calls; want the request served uncounted", err, backend.calls)
	}
}
