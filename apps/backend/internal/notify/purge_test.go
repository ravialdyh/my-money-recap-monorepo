package notify

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// members.deleted purges every row notify holds for the Member and stages notify.member.purged in its outbox, safely
// twice, the second replacing the first event, while another Member's rows stay (ARCHITECTURE §7.1 step 5).
func TestMembersDeletedPurgesEveryRow(t *testing.T) {
	db := scheduleDB(t)
	ctx := context.Background()
	gone, kept := "00000000-0000-4000-8000-00000000f011", "00000000-0000-4000-8000-00000000f012"
	for i, id := range []string{gone, kept} {
		for _, insert := range []string{
			`INSERT INTO member_facts (member_id, keycloak_subject, timezone, email) VALUES ($1, 's', 'Europe/London', 'x@example.test')`,
			`INSERT INTO notification_preference VALUES ($1, true, true, '20:00', true)`,
			`INSERT INTO device (apns_token, member_id) VALUES ('aa` + string(rune('0'+i)) + `', $1)`,
			`INSERT INTO month_flag (member_id, year_month, target_warning_message) VALUES ($1, '2026-09', 'over')`,
			`INSERT INTO sent_log (member_id, kind, period) VALUES ($1, 'recap_ready', '2026-09')`,
			`INSERT INTO idempotency (member_id, idempotency_key, request_hash, response) VALUES ($1, 'k', ''::bytea, ''::bytea)`,
		} {
			if _, err := db.Exec(ctx, insert, id); err != nil {
				t.Fatal(err)
			}
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
	if err := db.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE member_id = $1 AND routing_key = 'notify.member.purged'`, gone).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if rows := held(gone); events != 1 || rows != events {
		t.Errorf("after two deliveries notify holds %d rows for the Member, %d of them purged events; want only the one event", rows, events)
	}
	if rows := held(kept); rows != 6 {
		t.Errorf("another Member's rows: %d, want all six kept", rows)
	}
}
