package ledger

import (
	"errors"
	"math/big"
	"strings"
	"unicode"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/cldr"
)

// errNotCurrency is a code no territory takes as tender in CLDR.
var errNotCurrency = errors.New("currency_code is not an ISO 4217 currency")

// cash is one Currency as ledger holds it: its ISO 4217 code and the decimals CLDR gives it (package cldr, §7.2
// "Rounding"): GBP 2, IDR 0. The phone prints an amount with exactly these decimals.
type cash struct {
	code  string
	scale int
}

func currencyOf(code string) (cash, error) {
	scale, ok := cldr.Digits(code)
	if !ok {
		return cash{}, errNotCurrency
	}
	return cash{code: code, scale: scale}, nil
}

// money is r at the currency's decimals, halves rounded up (PRD §8.4): big.Rat rounds half away from zero,
// which is half up for every amount ledger holds, since none is negative.
func (c cash) money(r *big.Rat) *moneyrecapv1.Money {
	return &moneyrecapv1.Money{Amount: r.FloatString(c.scale), CurrencyCode: c.code}
}

// rounded is r at the currency's decimals, as money stores it.
func (c cash) rounded(r *big.Rat) *big.Rat {
	v, _ := new(big.Rat).SetString(r.FloatString(c.scale))
	return v
}

// printed is an amount in a sentence as the PRD prints one (§8.7): the home locale's symbol and grouping, a space
// only after a letter, no fraction when whole: "£320", "Rp 400.000"; the code where x/text lacks one: "VES 1.500".
func (c cash) printed(r *big.Rat) string {
	region, _ := language.ParseRegion(c.code[:2])
	base, _ := language.Make("und-" + region.String()).Base()
	tag, _ := language.Compose(base, region)
	p := message.NewPrinter(tag)
	symbol := c.code
	if u, err := currency.ParseISO(c.code); err == nil {
		symbol = p.Sprint(currency.Symbol(u))
	}
	digits := c.scale
	if r.IsInt() || c.rounded(r).IsInt() {
		digits = 0
	}
	f, _ := c.rounded(r).Float64()
	amount := p.Sprint(number.Decimal(f, number.Scale(digits)))
	last := []rune(symbol)[len([]rune(symbol))-1]
	if unicode.IsLetter(last) {
		return symbol + " " + amount
	}
	return strings.TrimSpace(symbol) + amount
}
