package ledger

import (
	"context"
	"math/big"
	"slices"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// ListSpends answers one month's Spend list (PRD §6.10): its Spends newest first, its Category rows and Spent.
func (s *Service) ListSpends(ctx context.Context, req *connect.Request[moneyrecapv1.ListSpendsRequest]) (*connect.Response[moneyrecapv1.ListSpendsResponse], error) {
	m, err := parseMonth(req.Msg.Month)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	spent, _ := l.spent(m)
	res := &moneyrecapv1.ListSpendsResponse{Spent: l.cash.money(spent), Categories: l.split(m, spent)}
	for _, sp := range slices.Backward(l.spends) {
		if monthOf(sp.date) == m {
			res.Spends = append(res.Spends, sp.spend)
		}
	}
	return connect.NewResponse(res), nil
}

// ListMonthsWithData answers the month switcher's list (PRD D59): the current month, then pastMonths.
func (s *Service) ListMonthsWithData(ctx context.Context, req *connect.Request[moneyrecapv1.ListMonthsWithDataRequest]) (*connect.Response[moneyrecapv1.ListMonthsWithDataResponse], error) {
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	months := []string{monthOf(l.today).String()}
	for _, m := range l.pastMonths(l.today) {
		months = append(months, m.String())
	}
	return connect.NewResponse(&moneyrecapv1.ListMonthsWithDataResponse{Months: months}), nil
}

// pastMonths is every month before today's with a Spend or Money In, newest first (PRD D59, §8.11): Home's
// past-Recap list, and the Spend list's switcher after the current month.
func (l ledgerOf) pastMonths(today time.Time) []month {
	first := monthOf(today)
	for _, s := range l.spends {
		if monthOf(s.date).before(first) {
			first = monthOf(s.date)
		}
	}
	for _, s := range l.salaries {
		if s.from.before(first) {
			first = s.from
		}
	}
	var list []month
	for m := monthOf(today).prev(); !m.before(first); m = m.prev() {
		if l.hasData(m, today) {
			list = append(list, m)
		}
	}
	return list
}

// split is month m's Category rows (PRD §6.10, §6.13): each Category with a Spend, its total, its share of spent
// and its count, largest first, then in D20's order.
func (l ledgerOf) split(m month, spent *big.Rat) []*moneyrecapv1.CategoryShare {
	shares := map[moneyrecapv1.Category]*moneyrecapv1.CategoryShare{}
	sums := map[moneyrecapv1.Category]*big.Rat{}
	for _, sp := range l.spends {
		if monthOf(sp.date) != m {
			continue
		}
		c := sp.spend.Category
		if shares[c] == nil {
			shares[c], sums[c] = &moneyrecapv1.CategoryShare{Category: c}, new(big.Rat)
		}
		shares[c].SpendCount++
		sums[c].Add(sums[c], sp.amount)
	}
	var rows []*moneyrecapv1.CategoryShare
	for c, share := range shares {
		share.Amount, share.Percent = l.cash.money(sums[c]), percentOf(sums[c], spent)
		rows = append(rows, share)
	}
	slices.SortFunc(rows, func(a, b *moneyrecapv1.CategoryShare) int {
		if n := sums[b.Category].Cmp(sums[a.Category]); n != 0 {
			return n
		}
		return int(a.Category - b.Category)
	})
	return rows
}

// hasData is true when month m has a Spend or Money In (PRD D59, §8.11), the test for a past month's Recap.
func (l ledgerOf) hasData(m month, today time.Time) bool {
	_, count := l.spent(m)
	return count > 0 || l.summary(m, today).Saved != nil
}
