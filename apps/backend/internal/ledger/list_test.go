package ledger

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
)

// The switcher lists the current month and every past month with a Spend or Money In, newest first, as Home's list
// does; a month lists its Spends newest first, its Categories largest first with share and count (PRD §6.10, D59).
func TestTheSpendListAndItsMonths(t *testing.T) {
	db := testDB(t)
	id := "00000000-0000-4000-8000-00000000a701"
	gbp := func(amount string) *moneyrecapv1.Money {
		return &moneyrecapv1.Money{Amount: amount, CurrencyCode: "GBP"}
	}
	ctx := seed(t, db, id, "lister", "Europe/London", "2026-04-01", "GBP",
		[]*moneyrecapv1.SalarySetting{{Amount: gbp("2400.00"), OriginalAmount: gbp("2400.00"), Payday: "25", EffectiveFrom: "2026-06"}}, nil)
	var rows []*moneyrecapv1.Spend
	for _, s := range []struct {
		amount, date, note string
		c                  moneyrecapv1.Category
	}{
		{"50.00", "2026-04-03", "", 1}, {"1500.00", "2026-07-05", "Lego", 3}, {"500.00", "2026-07-10", "", 1}, {"20.00", "2026-07-10", "Kopi", 1},
	} {
		rows = append(rows, &moneyrecapv1.Spend{Amount: gbp(s.amount), OriginalAmount: gbp(s.amount), Category: s.c, Date: s.date, Note: s.note})
	}
	seedSpends(t, db, id, rows)
	s := &Service{DB: db, Now: func() time.Time { return at(t, "2026-09-15", "Europe/London") }}

	months, err := s.ListMonthsWithData(ctx, connect.NewRequest(&moneyrecapv1.ListMonthsWithDataRequest{MemberId: id}))
	if want := []string{"2026-09", "2026-08", "2026-07", "2026-06", "2026-04"}; err != nil || !slices.Equal(months.Msg.Months, want) {
		t.Errorf("months %v, %v, want %v: May has nothing, September is the current month", months, err, want)
	}
	list := func(month string) (*moneyrecapv1.ListSpendsResponse, error) {
		res, err := s.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.ListSpendsRequest{MemberId: id, Month: month}))
		if err != nil {
			return nil, err
		}
		return res.Msg, nil
	}
	july, err := list("2026-07")
	if err != nil {
		t.Fatal(err)
	}
	var notes []string
	for _, sp := range july.Spends {
		notes = append(notes, sp.Date+" "+sp.Amount.Amount+" "+sp.Note)
	}
	if want := []string{"2026-07-10 20.00 Kopi", "2026-07-10 500.00 ", "2026-07-05 1500.00 Lego"}; !slices.Equal(notes, want) {
		t.Errorf("July's Spends %q, want %q, newest first", notes, want)
	}
	if c := july.Categories; july.Spent.Amount != "2020.00" || len(c) != 2 || c[0].Category != moneyrecapv1.Category_CATEGORY_TOYS ||
		c[0].Percent != 74 || c[0].SpendCount != 1 || c[1].Amount.Amount != "520.00" || c[1].Percent != 26 || c[1].SpendCount != 2 {
		t.Errorf("July: %v, want Toys £1,500 at 74%% then Snacks £520 at 26%% in two", july)
	}
	if may, err := list("2026-05"); err != nil || may.Spent.Amount != "0.00" || len(may.Spends)+len(may.Categories) != 0 {
		t.Errorf("May: %v, %v, want nothing at £0.00", may, err)
	}
	if _, err := list("July"); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("a month that is not YYYY-MM: %v, want invalid_argument", err)
	}
	withMembers(t, s, nil)
	stranger := auth.NewContext(context.Background(), auth.Claims{Subject: "stranger", Raw: "stranger"})
	if _, err := s.ListSpends(stranger, connect.NewRequest(&moneyrecapv1.ListSpendsRequest{MemberId: id, Month: "2026-07"})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("another token: %v, want not_found", err)
	}
}

// Spent today is each Member's own today: at 20:00 UTC Jakarta is on the 16th, London on the 15th. notify's token
// reads both and the month list; unknown ids are left out; a Member's token reads only its own (§4.2).
func TestSpentTodayIsEachMembersOwnToday(t *testing.T) {
	db := testDB(t)
	jakarta, london := "00000000-0000-4000-8000-00000000a771", "00000000-0000-4000-8000-00000000a772"
	idrSalary := []*moneyrecapv1.SalarySetting{{Amount: idr("15000000"), OriginalAmount: idr("15000000"), Payday: "1", EffectiveFrom: "2026-09"}}
	owner := seed(t, db, jakarta, "jakarta", "Asia/Jakarta", "2026-09-01", "IDR", idrSalary, nil)
	seed(t, db, london, "london", "Europe/London", "2026-09-01", "GBP", nil, nil)
	seedSpends(t, db, jakarta, []*moneyrecapv1.Spend{
		{Amount: idr("25000"), OriginalAmount: idr("25000"), Category: 1, Date: "2026-09-16"},
		{Amount: idr("40000"), OriginalAmount: idr("40000"), Category: 2, Date: "2026-09-16"},
		{Amount: idr("9000"), OriginalAmount: idr("9000"), Category: 1, Date: "2026-09-15"},
	})
	s := &Service{DB: db, Now: func() time.Time { return time.Date(2026, 9, 15, 20, 0, 0, 0, time.UTC) }}
	withMembers(t, s, nil)
	job := auth.NewContext(context.Background(), auth.Claims{Subject: "service-account-notify", Job: true})
	ids := []string{london, "00000000-0000-4000-8000-00000000a779", "not-a-uuid", jakarta}

	res, err := s.GetSpentToday(job, connect.NewRequest(&moneyrecapv1.GetSpentTodayRequest{MemberIds: ids}))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range res.Msg.Entries {
		got = append(got, fmt.Sprintf("%s %s %s %d", e.MemberId[len(e.MemberId)-4:], e.Spent.Amount, e.Spent.CurrencyCode, e.SpendCount))
	}
	if want := []string{"a772 0 GBP 0", "a771 65000 IDR 2"}; !slices.Equal(got, want) {
		t.Errorf("spent today %q, want %q", got, want)
	}
	months, err := s.ListMonthsWithData(job, connect.NewRequest(&moneyrecapv1.ListMonthsWithDataRequest{MemberId: jakarta}))
	if err != nil || !slices.Equal(months.Msg.Months, []string{"2026-09"}) {
		t.Errorf("notify's month list: %v, %v, want September", months, err)
	}

	own, err := s.GetSpentToday(owner, connect.NewRequest(&moneyrecapv1.GetSpentTodayRequest{MemberIds: ids}))
	if err != nil || len(own.Msg.Entries) != 1 || own.Msg.Entries[0].MemberId != jakarta {
		t.Errorf("a Member's own token: %v, %v, want its own entry alone", own, err)
	}
}
