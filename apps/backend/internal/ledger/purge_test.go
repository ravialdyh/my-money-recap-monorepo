package ledger

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/testdb"
)

// members.deleted purges every row ledger holds for the Member and stages ledger.member.purged, safely twice, the
// second replacing the first event, while another Member's rows stay (ARCHITECTURE §7.1 step 5).
func TestMembersDeletedPurgesEveryRow(t *testing.T) {
	db := testdb.Scratch(t, "LEDGER_TEST_DATABASE_URL", Migrate)
	ctx := context.Background()
	gone, kept := "00000000-0000-4000-8000-00000000f001", "00000000-0000-4000-8000-00000000f002"
	money := &moneyrecapv1.Money{Amount: "10", CurrencyCode: "GBP"}
	for _, id := range []string{gone, kept} {
		seed(t, db, id, "s-"+id, "Europe/London", "2026-09-01", "GBP",
			[]*moneyrecapv1.SalarySetting{{Amount: money, OriginalAmount: money, Payday: "1", EffectiveFrom: "2026-09"}},
			[]*moneyrecapv1.SavingTargetSetting{{Percent: 20, EffectiveFrom: "2026-09"}})
		seedSpends(t, db, id, []*moneyrecapv1.Spend{{Amount: money, OriginalAmount: money, Category: 1, Date: "2026-09-02"}})
		if _, err := db.Exec(ctx, `INSERT INTO idempotency (member_id, idempotency_key, request_hash, response) VALUES ($1, 'k', ''::bytea, ''::bytea)`, id); err != nil {
			t.Fatal(err)
		}
	}
	body, err := proto.Marshal(&moneyrecapv1.MembersDeleted{Envelope: &moneyrecapv1.EventEnvelope{MemberId: gone}})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
			return ApplyMemberEvent(ctx, tx, amqp.Delivery{RoutingKey: "members.deleted", Body: body})
		}); err != nil {
			t.Fatal(err)
		}
	}
	held := func(id string) int {
		t.Helper()
		total := 0
		for _, table := range purged {
			var n int
			if err := db.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE member_id = $1`, id).Scan(&n); err != nil {
				t.Fatal(err)
			}
			total += n
		}
		return total
	}
	var events int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE member_id = $1 AND routing_key = 'ledger.member.purged'`, gone).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if rows := held(gone); events != 1 || rows != events {
		t.Errorf("after two deliveries ledger holds %d rows for the Member, %d of them purged events; want only the one event", rows, events)
	}
	if rows := held(kept); rows != 5 {
		t.Errorf("another Member's rows: %d, want all five kept", rows)
	}
}
