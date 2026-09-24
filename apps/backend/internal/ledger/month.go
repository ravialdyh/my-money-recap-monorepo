package ledger

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// month is a Recap Month (PRD §8.1): a calendar month of the Member's local dates.
type month struct {
	year int
	mon  time.Month
}

func monthOf(d time.Time) month { return month{d.Year(), d.Month()} }

func parseMonth(s string) (month, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return month{}, fmt.Errorf("month %q is not YYYY-MM", s)
	}
	return monthOf(t), nil
}

func (m month) String() string      { return fmt.Sprintf("%04d-%02d", m.year, m.mon) }
func (m month) day(d int) time.Time { return time.Date(m.year, m.mon, d, 0, 0, 0, 0, time.UTC) }
func (m month) days() int           { return m.day(1).AddDate(0, 1, -1).Day() }
func (m month) next() month         { return monthOf(m.day(1).AddDate(0, 1, 0)) }
func (m month) prev() month         { return monthOf(m.day(1).AddDate(0, -1, 0)) }
func (m month) before(o month) bool { return m.day(1).Before(o.day(1)) }

// effectivePayday is PRD §8.2: the token "last" is the month's last day, and so is a Payday past it.
func effectivePayday(payday string, m month) int {
	if payday == "last" {
		return m.days()
	}
	n, _ := strconv.Atoi(payday)
	return min(n, m.days())
}

// salaryRow and targetRow are the append-only settings rows (PRD §7), read oldest first.
type salaryRow struct {
	amount, original *big.Rat
	originalCurrency string
	payday           string
	from             month
}

type targetRow struct {
	percent int32
	from    month
}

// inForce is the row with the latest effective month not after m (PRD §7). Rows come oldest first, so of two
// sharing that month the newer wins, which makes a step 3 save repeated after a crash safe (spec #92).
func inForce[T any](rows []T, from func(T) month, m month) (T, bool) {
	var found T
	ok := false
	for _, r := range rows {
		if !m.before(from(r)) && (!ok || !from(r).before(from(found))) {
			found, ok = r, true
		}
	}
	return found, ok
}

func salaryFrom(r salaryRow) month { return r.from }
func targetFrom(r targetRow) month { return r.from }

// spendRow is a Spend that is not deleted (PRD §7): its amount and local date as every figure reads them, and the
// Spend as ledger answers it.
type spendRow struct {
	amount *big.Rat
	date   time.Time
	spend  *moneyrecapv1.Spend
}

// ledgerOf is one Member's rows and replica facts, what every computation reads.
type ledgerOf struct {
	cash     cash
	signedUp time.Time
	salaries []salaryRow
	targets  []targetRow
	spends   []spendRow
}

// spent is PRD §8.3's Spent for month m and how many Spends make it.
func (l ledgerOf) spent(m month) (sum *big.Rat, count int) {
	sum = new(big.Rat)
	for _, s := range l.spends {
		if monthOf(s.date) == m {
			sum.Add(sum, s.amount)
			count++
		}
	}
	return sum, count
}

// summary is GetMonthSummary for month m on the Member's local today (PRD §8): the figures, the §8.7 sentence
// and the §8.8 warning copy, present whenever Spent is above the line and worded by whether Payday has passed (D53).
func (l ledgerOf) summary(m month, today time.Time) *moneyrecapv1.MonthSummary {
	spent, count := l.spent(m)
	s := &moneyrecapv1.MonthSummary{
		Month: m.String(), Spent: l.cash.money(spent), MoneyIn: l.cash.money(new(big.Rat)),
		StreakDay: l.streak(m, today), DaysRemaining: daysRemaining(m, today),
	}
	sal, ok := inForce(l.salaries, salaryFrom, m)
	if !ok {
		s.Partial = m.before(monthOf(today)) && count > 0
		return s
	}
	payday := m.day(effectivePayday(sal.payday, m))
	current := m == monthOf(today)
	s.Partial = m == monthOf(l.signedUp) && payday.Before(l.signedUp)
	paid := !s.Partial && (!current || !today.Before(payday))
	saved := new(big.Rat).Sub(sal.amount, spent)
	if paid {
		s.MoneyIn, s.Saved = l.cash.money(sal.amount), l.cash.money(saved)
	}
	s.Salary = &moneyrecapv1.SalaryFacts{Amount: l.cash.money(sal.amount), Payday: sal.payday}
	switch {
	case current && today.Before(payday):
		s.Salary.DaysUntilPayday = int32(payday.Sub(today).Hours() / 24)
	case current && s.Partial:
		s.Salary.DaysUntilPayday = int32(m.next().day(effectivePayday(sal.payday, m.next())).Sub(today).Hours() / 24)
	}
	tgt, ok := inForce(l.targets, targetFrom, m)
	if !ok || tgt.percent == 0 {
		return s
	}
	target := l.cash.rounded(new(big.Rat).Mul(sal.amount, big.NewRat(int64(tgt.percent), 100)))
	line := new(big.Rat).Sub(sal.amount, target)
	over := new(big.Rat).Sub(spent, line)
	s.SavingTarget = &moneyrecapv1.SavingTargetFacts{Percent: tgt.percent, Amount: l.cash.money(target)}
	spare := new(big.Rat).Sub(saved, target)
	switch {
	case !paid && (!m.before(monthOf(today)) || !s.Partial):
		next := payday
		if s.Partial {
			next = m.next().day(effectivePayday(sal.payday, m.next()))
		}
		s.SavingTarget.ProgressSentence = "Nothing to keep yet. Your salary lands on the " + ordinal(next.Day()) + "."
	case current && over.Sign() > 0:
		s.SavingTarget.ProgressSentence = "You're " + l.cash.printed(over) + " past your line. Next month is a fresh start."
	case current:
		left := new(big.Rat).Neg(over)
		daily := new(big.Int).Quo(left.Num(), new(big.Int).Mul(left.Denom(), big.NewInt(int64(s.DaysRemaining))))
		s.SavingTarget.ProgressSentence = "Keep spending under " + l.cash.printed(new(big.Rat).SetInt(daily)) + " a day and you'll make it."
	case paid && spare.Sign() >= 0:
		s.SavingTarget.ProgressSentence = "Target met. " + l.cash.printed(spare) + " to spare."
	case paid:
		s.SavingTarget.ProgressSentence = l.cash.printed(new(big.Rat).Neg(spare)) + " short of your target."
	}
	switch {
	case over.Sign() <= 0:
	case current && today.Before(payday):
		s.TargetWarning = &moneyrecapv1.TargetWarning{Message: "You've spent " + l.cash.printed(spent) + " of the " + l.cash.printed(line) + " you meant to spend this month."}
	default:
		s.TargetWarning = &moneyrecapv1.TargetWarning{Message: "You've passed your saving line. " + l.cash.printed(over) + " over."}
	}
	return s
}

// streak is PRD §8.5: the longest run of spend-free days in the window, which starts on the later of the 1st and
// the sign-up date and ends today in the current month or on the month's last day before it.
func (l ledgerOf) streak(m month, today time.Time) int32 {
	start, end := l.window(m, today)
	spendDays := map[time.Time]bool{}
	for _, s := range l.spends {
		spendDays[s.date] = true
	}
	var best, run int32
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		run++
		if spendDays[d] {
			run = 0
		}
		best = max(best, run)
	}
	return best
}

// window is the Streak window of month m on today (PRD §8.5, D43).
func (l ledgerOf) window(m month, today time.Time) (start, end time.Time) {
	start, end = m.day(1), m.day(m.days())
	if l.signedUp.After(start) {
		start = l.signedUp
	}
	if m == monthOf(today) {
		end = today
	}
	return start, end
}

// daysRemaining counts today and is never below 1 (PRD §8.7), so a past month answers 1.
func daysRemaining(m month, today time.Time) int32 {
	switch {
	case m == monthOf(today):
		return int32(m.days() - today.Day() + 1)
	case m.before(monthOf(today)):
		return 1
	}
	return int32(m.days())
}

// ordinal is an English day of the month as the §8.7 sentence prints it: 1st, 2nd, 3rd, 11th, 22nd, 25th.
func ordinal(d int) string {
	suffix := "th"
	if d%100 < 11 || d%100 > 13 {
		switch d % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(d) + suffix
}

// settings is GetSettings on the Member's local today (PRD §6.16): the rows in force this month, any saved
// for a later month, and whether this month's Payday has come, which opens the D56 sheet.
func (l ledgerOf) settings(today time.Time) *moneyrecapv1.LedgerSettings {
	now := monthOf(today)
	st := &moneyrecapv1.LedgerSettings{CurrencyCode: l.cash.code, HasAmountsToConvert: len(l.salaries)+len(l.spends) > 0}
	if sal, ok := inForce(l.salaries, salaryFrom, now); ok {
		st.Salary = l.salarySetting(sal)
		st.PaydayPassedThisMonth = !today.Before(now.day(effectivePayday(sal.payday, now)))
	}
	if sal, ok := inForce(l.salaries, salaryFrom, farFuture); ok && now.before(sal.from) {
		st.NextSalary = l.salarySetting(sal)
	}
	if tgt, ok := inForce(l.targets, targetFrom, now); ok && tgt.percent > 0 {
		st.SavingTarget = &moneyrecapv1.SavingTargetSetting{Percent: tgt.percent, EffectiveFrom: tgt.from.String()}
	}
	if tgt, ok := inForce(l.targets, targetFrom, farFuture); ok && now.before(tgt.from) {
		st.NextSavingTarget = &moneyrecapv1.SavingTargetSetting{Percent: tgt.percent, EffectiveFrom: tgt.from.String()}
	}
	return st
}

// farFuture is a month after every effective month, so the row in force then is the last one saved.
var farFuture = month{9999, time.December}

func (l ledgerOf) salarySetting(r salaryRow) *moneyrecapv1.SalarySetting {
	original, _ := currencyOf(r.originalCurrency)
	return &moneyrecapv1.SalarySetting{
		Amount: l.cash.money(r.amount), Payday: r.payday, EffectiveFrom: r.from.String(), OriginalAmount: original.money(r.original),
	}
}
