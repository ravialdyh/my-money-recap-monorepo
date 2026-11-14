package notify

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// TestNotifysKeyReadsTheRealLedgerForABatch ticks both schedules for two throwaway Members against the ledger
// container with notify's own token: the one with a Spend today and one last month hears the Recap alone, the other
// the reminder alone.
func TestNotifysKeyReadsTheRealLedgerForABatch(t *testing.T) {
	base, ledgerDSN, kc, secret := os.Getenv("LEDGER_TEST_URL"), os.Getenv("LEDGER_TEST_DATABASE_URL"), os.Getenv("MMR_TEST_KEYCLOAK"), os.Getenv("NOTIFY_TEST_CLIENT_SECRET")
	if base == "" || ledgerDSN == "" || kc == "" || secret == "" {
		t.Skip("LEDGER_TEST_URL, LEDGER_TEST_DATABASE_URL, MMR_TEST_KEYCLOAK and NOTIFY_TEST_CLIENT_SECRET not all set")
	}
	ledgerDB, err := pgxpool.New(t.Context(), ledgerDSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ledgerDB.Close)
	busy, idle := "00000000-0000-4000-8000-00000000d401", "00000000-0000-4000-8000-00000000d402"
	t.Cleanup(func() {
		for _, table := range []string{"spend", "member_facts"} {
			if _, err := ledgerDB.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = ANY($1)`, []string{busy, idle}); err != nil {
				t.Logf("cleanup ledger %s: %v", table, err)
			}
		}
	})
	wib := time.FixedZone("WIB", 7*60*60)
	today := time.Now().In(wib)
	lastMonth := time.Date(today.Year(), today.Month()-1, 1, 0, 0, 0, 0, time.UTC)
	for _, id := range []string{busy, idle} {
		if _, err := ledgerDB.Exec(t.Context(), `INSERT INTO member_facts (member_id, timezone, signed_up_on, currency_code) VALUES ($1, 'Asia/Jakarta', $2, 'IDR')`, id, lastMonth.AddDate(0, -1, 0)); err != nil {
			t.Fatal(err)
		}
	}
	for _, day := range []time.Time{today, lastMonth} {
		if _, err := ledgerDB.Exec(t.Context(), `INSERT INTO spend (member_id, amount, original_amount, original_currency, category, spent_on) VALUES ($1, 12000, 12000, 'IDR', 1, $2)`, busy, day.Format(time.DateOnly)); err != nil {
			t.Fatal(err)
		}
	}

	account := &auth.ServiceAccount{Realm: kc + "/realms/mmr", Client: "notify", Secret: secret}
	apple, fake, _ := startApple(t)
	s := &Service{DB: scheduleDB(t), Ledger: moneyrecapv1connect.NewLedgerServiceClient(http.DefaultClient, base, connect.WithInterceptors(account.Bearer())), Apple: apple}
	now := time.Now()
	for _, id := range []string{busy, idle} {
		scheduled(t, s, id, "Asia/Jakarta", "20:00", now)
		if _, err := s.DB.Exec(t.Context(), `UPDATE member_facts SET next_recap_at = $2, next_reminder_at = $2 WHERE member_id = $1`, id, now.Add(-time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	s.tick(t.Context(), now)
	got := alerts(t, fake)
	recap := "Your " + lastMonth.Format("January") + " recap is ready. See how you did."
	if len(got) != 2 || len(got[deviceOf(busy)]) != 1 || got[deviceOf(busy)][0] != recap || len(got[deviceOf(idle)]) != 1 || got[deviceOf(idle)][0] != "Log today's spends before you forget." {
		t.Fatalf("Apple heard %q, want the Recap for the Member with Spends and the reminder for the other", got)
	}
	t.Logf("ledger answered notify's key for a batch of two: %q", got)
}
