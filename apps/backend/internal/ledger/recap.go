package ledger

import (
	"context"
	"errors"
	"math/big"
	"slices"
	"strconv"
	"time"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// GetRecap answers one month's Recap on the Member's local today; a month after the current one has none yet.
func (s *Service) GetRecap(ctx context.Context, req *connect.Request[moneyrecapv1.GetRecapRequest]) (*connect.Response[moneyrecapv1.GetRecapResponse], error) {
	m, err := parseMonth(req.Msg.Month)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	if monthOf(l.today).before(m) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("month is after the current month"))
	}
	return connect.NewResponse(&moneyrecapv1.GetRecapResponse{Recap: l.recap(m, l.today)}), nil
}

// archetypes are PRD §8.6's names and one-line descriptions, the phone's only source for either (A2).
var archetypes = map[moneyrecapv1.Archetype][2]string{
	moneyrecapv1.Archetype_ARCHETYPE_SQUIRREL:       {"Squirrel", "You stashed more than you meant to, and you didn't crack once."},
	moneyrecapv1.Archetype_ARCHETYPE_HUSTLER:        {"Hustler", "You hit your number while still living a little."},
	moneyrecapv1.Archetype_ARCHETYPE_TREAT_YOURSELF: {"Treat-Yourself", "One big treat tipped the month. Worth it?"},
	moneyrecapv1.Archetype_ARCHETYPE_SLOW_BURNER:    {"Slow Burner", "Lots of little spends added up. Sneaky."},
}

// recap is month m's Recap on today (PRD §8, §6.11 to §6.14): the current month is the preview and a partial month
// (§8.11) is in its situation, each with its §6.17 closing line and no Archetype.
func (l ledgerOf) recap(m month, today time.Time) *moneyrecapv1.Recap {
	s := l.summary(m, today)
	r := &moneyrecapv1.Recap{
		Month: m.String(), MoneyIn: s.MoneyIn, Spent: s.Spent, Saved: s.Saved, Salary: s.Salary, SavingTarget: s.SavingTarget,
		Streak: l.streakFacts(m, today),
	}
	switch {
	case m == monthOf(today):
		r.Situation = moneyrecapv1.PartialSituation_PARTIAL_SITUATION_CURRENT_MONTH
		r.ClosingLine = "Your recap arrives on " + dayMonth(m.next().day(1)) + "."
	case s.Partial && s.Salary != nil:
		r.Situation = moneyrecapv1.PartialSituation_PARTIAL_SITUATION_SIGN_UP_MONTH
		r.ClosingLine = "Your first full recap arrives on " + dayMonth(m.next().next().day(1)) + "."
	case s.Partial:
		r.Situation = moneyrecapv1.PartialSituation_PARTIAL_SITUATION_NO_MONEY_IN
		r.ClosingLine = "Set your monthly salary to get a full recap."
	}
	spent, _ := l.spent(m)
	shares := map[moneyrecapv1.Category]*moneyrecapv1.CategoryShare{}
	sums := map[moneyrecapv1.Category]*big.Rat{}
	var splurge spendRow
	for _, sp := range l.spends {
		if monthOf(sp.date) != m {
			continue
		}
		r.SpendCount++
		c := sp.spend.Category
		if shares[c] == nil {
			shares[c], sums[c] = &moneyrecapv1.CategoryShare{Category: c}, new(big.Rat)
		}
		shares[c].SpendCount++
		sums[c].Add(sums[c], sp.amount)
		if splurge.spend == nil || sp.amount.Cmp(splurge.amount) > 0 {
			splurge = sp
		}
	}
	for c, share := range shares {
		share.Amount, share.Percent = l.cash.money(sums[c]), percentOf(sums[c], spent)
		r.Categories = append(r.Categories, share)
	}
	slices.SortFunc(r.Categories, func(a, b *moneyrecapv1.CategoryShare) int {
		if n := sums[b.Category].Cmp(sums[a.Category]); n != 0 {
			return n
		}
		return int(a.Category - b.Category)
	})
	r.Splurge = splurge.spend
	if s.Saved == nil || s.SavingTarget == nil {
		return r
	}
	saved, _ := new(big.Rat).SetString(s.Saved.Amount)
	target, _ := new(big.Rat).SetString(s.SavingTarget.Amount.Amount)
	kept := min(max(percentOf(saved, target), 0), 999)
	r.KeptPercent = &kept
	if r.Situation != moneyrecapv1.PartialSituation_PARTIAL_SITUATION_UNSPECIFIED {
		return r
	}
	a := moneyrecapv1.Archetype_ARCHETYPE_SLOW_BURNER
	switch {
	case saved.Cmp(target) >= 0 && r.Streak.Days >= 7:
		a = moneyrecapv1.Archetype_ARCHETYPE_SQUIRREL
	case saved.Cmp(target) >= 0:
		a = moneyrecapv1.Archetype_ARCHETYPE_HUSTLER
	case new(big.Rat).Mul(splurge.amount, big.NewRat(4, 1)).Cmp(spent) >= 0:
		a = moneyrecapv1.Archetype_ARCHETYPE_TREAT_YOURSELF
	}
	r.Archetype, r.ArchetypeName, r.ArchetypeDescription = &a, archetypes[a][0], archetypes[a][1]
	return r
}

// percentOf is round(part ÷ whole × 100), halves away from zero (§8.4, §6.13), held at int32's bounds; 0 when whole is 0.
func percentOf(part, whole *big.Rat) int32 {
	if whole.Sign() == 0 {
		return 0
	}
	n, _ := strconv.ParseInt(new(big.Rat).Quo(new(big.Rat).Mul(part, big.NewRat(100, 1)), whole).FloatString(0), 10, 32)
	return int32(n)
}

// streakFacts is the Streak with its window and one spend-free flag per day (PRD §8.5, §6.12).
func (l ledgerOf) streakFacts(m month, today time.Time) *moneyrecapv1.StreakFacts {
	start, end := l.window(m, today)
	f := &moneyrecapv1.StreakFacts{Days: l.streak(m, today), WindowStart: start.Format(time.DateOnly), WindowEnd: end.Format(time.DateOnly)}
	spendDays := map[time.Time]bool{}
	for _, s := range l.spends {
		spendDays[s.date] = true
	}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		f.SpendFreeDays = append(f.SpendFreeDays, !spendDays[d])
	}
	f.BadgeUnlocked = f.Days >= 7
	return f
}

// dayMonth is a date as the §6.17 closing line prints it: "1 November".
func dayMonth(d time.Time) string { return strconv.Itoa(d.Day()) + " " + d.Month().String() }

// recapMonths is Home's past-Recap list on today (PRD §6.8, §8.11): every month before the current one with a Spend
// or Money In, newest first, labelled by its Archetype name or "Partial".
func (l ledgerOf) recapMonths(today time.Time) []*moneyrecapv1.RecapMonth {
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
	var list []*moneyrecapv1.RecapMonth
	for m := monthOf(today).prev(); !m.before(first); m = m.prev() {
		r := l.recap(m, today)
		if r.SpendCount == 0 && r.Saved == nil {
			continue
		}
		label := r.ArchetypeName
		if label == "" {
			label = "Partial"
		}
		list = append(list, &moneyrecapv1.RecapMonth{Month: m.String(), Label: label})
	}
	return list
}
