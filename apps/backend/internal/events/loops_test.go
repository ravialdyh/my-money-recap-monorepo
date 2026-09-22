package events_test

import (
	"context"
	"crypto/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/events"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/ledger"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/members"
)

// Relay polls the outbox and publishes what a write staged after it started, until its context ends.
func TestRelayPublishesWhatIsStaged(t *testing.T) {
	db := scratch(t, "MMR_TEST_MEMBERS_DATABASE_URL", members.Migrate)
	url, conn := broker(t)
	key := "test.loop." + strings.ToLower(rand.Text())
	received := listen(t, conn, key)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		events.Relay(ctx, db, url)
		close(done)
	}()
	ids := stage(t, db, key, 3)
	seen := received(len(ids))
	cancel()
	<-done
	for _, id := range ids {
		if seen[id] != 1 {
			t.Errorf("row %s published %d times, want once", id, seen[id])
		}
	}
}

// A broker that cannot be reached is dialled again until the context ends, and then both loops return.
func TestLoopsReturnWhenTheirContextEnds(t *testing.T) {
	db := scratch(t, "LEDGER_TEST_DATABASE_URL", ledger.Migrate)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	nowhere := "amqp://nobody:nothing@127.0.0.1:1/"
	start := time.Now()
	var wg sync.WaitGroup
	wg.Go(func() { events.Relay(ctx, db, nowhere) })
	wg.Go(func() { events.Consume(ctx, db, nowhere, "test.nowhere", nil, nil) })
	wg.Wait()
	if waited := time.Since(start); waited > 2*time.Second {
		t.Errorf("the loops took %v to notice their context ended", waited)
	}
}
