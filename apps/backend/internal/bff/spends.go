package bff

import (
	"context"

	"connectrpc.com/connect"
	"golang.org/x/sync/errgroup"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// ListSpends is the Spend list (PRD §6.10, D59) for the month asked, or the current one in the Member's zone: ledger's
// month, the switcher's months and the month's Streak, asked at once, each required, each code passed through.
func (s *Service) ListSpends(ctx context.Context, req *connect.Request[moneyrecapv1.BffServiceListSpendsRequest]) (*connect.Response[moneyrecapv1.BffServiceListSpendsResponse], error) {
	profile, err := s.member(ctx)
	if err != nil {
		return nil, err
	}
	month := req.Msg.Month
	if month == "" {
		month = s.now().In(zone(profile)).Format(monthLayout)
	}
	id := profile.MemberId
	var (
		list    *connect.Response[moneyrecapv1.ListSpendsResponse]
		months  *connect.Response[moneyrecapv1.ListMonthsWithDataResponse]
		summary *connect.Response[moneyrecapv1.GetMonthSummaryResponse]
	)
	g, gctx := errgroup.WithContext(ctx)
	ask := func(call string, f func(ctx context.Context) error) {
		g.Go(func() error {
			ctx, cancel := context.WithTimeout(gctx, s.timeout())
			defer cancel()
			if err := f(ctx); err != nil {
				return requiredCallFailed(call, connect.CodeOf(err), err)
			}
			return nil
		})
	}
	ask("ledger.ListSpends", func(ctx context.Context) (err error) {
		list, err = s.Ledger.ListSpends(ctx, connect.NewRequest(&moneyrecapv1.ListSpendsRequest{MemberId: id, Month: month}))
		return err
	})
	ask("ledger.ListMonthsWithData", func(ctx context.Context) (err error) {
		months, err = s.Ledger.ListMonthsWithData(ctx, connect.NewRequest(&moneyrecapv1.ListMonthsWithDataRequest{MemberId: id}))
		return err
	})
	ask("ledger.GetMonthSummary", func(ctx context.Context) (err error) {
		summary, err = s.Ledger.GetMonthSummary(ctx, connect.NewRequest(&moneyrecapv1.GetMonthSummaryRequest{MemberId: id, Month: month}))
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.BffServiceListSpendsResponse{Spends: &moneyrecapv1.SpendListView{
		Month: month, CurrencyCode: profile.CurrencyCode, CurrencyDigits: profile.CurrencyDigits, Months: months.Msg.Months,
		Spent: list.Msg.Spent, Categories: list.Msg.Categories, Spends: list.Msg.Spends, StreakDay: summary.Msg.Summary.GetStreakDay(),
	}}), nil
}
