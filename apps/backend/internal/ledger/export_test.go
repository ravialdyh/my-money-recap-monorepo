package ledger

import (
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// The export is every Spend left, oldest first, each amount exact in the current Currency, quoted by RFC 4180, and
// any cell a spreadsheet would run as a formula made text; with no Spends it is the header alone (PRD §6.16).
func TestTheExportCSV(t *testing.T) {
	db := testDB(t)
	id, empty := "00000000-0000-4000-8000-00000000a182", "00000000-0000-4000-8000-00000000a183"
	gbp := func(amount string) *moneyrecapv1.Money {
		return &moneyrecapv1.Money{Amount: amount, CurrencyCode: "GBP"}
	}
	ctx := seed(t, db, id, "exporter", "Europe/London", "2026-04-01", "GBP", nil, nil)
	var rows []*moneyrecapv1.Spend
	for _, s := range []struct {
		amount, date, note string
		c                  moneyrecapv1.Category
	}{
		{"4.50", "2026-07-10", `Kopi, "large"`, 1}, {"1500.00", "2026-07-05", "=HYPERLINK(1)", 3}, {"12.00", "2026-08-01", "-2+3", 8},
		{"9.99", "2026-06-02", "gone", 2}, {"20.00", "2026-06-03", "@sum\nline two", 6},
	} {
		rows = append(rows, &moneyrecapv1.Spend{Amount: gbp(s.amount), OriginalAmount: gbp(s.amount), Category: s.c, Date: s.date, Note: s.note})
	}
	seedSpends(t, db, id, rows)
	if _, err := db.Exec(t.Context(), `UPDATE spend SET deleted_at = now() WHERE member_id = $1 AND note = 'gone'`, id); err != nil {
		t.Fatal(err)
	}
	s := &Service{DB: db, Now: func() time.Time { return at(t, "2026-09-15", "Europe/London") }}
	res, err := s.ExportCSV(ctx, connect.NewRequest(&moneyrecapv1.ExportCSVRequest{MemberId: id}))
	if err != nil {
		t.Fatal(err)
	}
	want := "date,category,amount,currency,note\r\n" +
		"2026-06-03,Bills,20.00,GBP,\"'@sum\r\nline two\"\r\n" +
		"2026-07-05,Toys,1500.00,GBP,'=HYPERLINK(1)\r\n" +
		"2026-07-10,Snacks,4.50,GBP,\"Kopi, \"\"large\"\"\"\r\n" +
		"2026-08-01,Other,12.00,GBP,'-2+3\r\n"
	if got := string(res.Msg.Csv); got != want {
		t.Errorf("the export:\n%q\nwant\n%q", got, want)
	}

	ctx = seed(t, db, empty, "nothing", "Europe/London", "2026-04-01", "GBP", nil, nil)
	res, err = s.ExportCSV(ctx, connect.NewRequest(&moneyrecapv1.ExportCSVRequest{MemberId: empty}))
	if err != nil || string(res.Msg.Csv) != "date,category,amount,currency,note\r\n" {
		t.Errorf("no Spends: %v, %v, want the header alone", res, err)
	}
}
