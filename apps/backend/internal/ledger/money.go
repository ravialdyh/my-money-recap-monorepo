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
)

// errNotCurrency is a code CLDR does not know as a currency.
var errNotCurrency = errors.New("currency_code is not an ISO 4217 currency")

// cash is one Currency as ledger holds it: its ISO 4217 code and the decimals CLDR gives it, the data the
// phone's formatter prints from (§7.2 "Rounding"): GBP 2, IDR 0.
type cash struct {
	code  string
	scale int
}

func currencyOf(code string) (cash, error) {
	u, err := currency.ParseISO(code)
	if err != nil || len(code) != 3 {
		return cash{}, errNotCurrency
	}
	scale, _ := currency.Standard.Rounding(u)
	return cash{code: u.String(), scale: scale}, nil
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

// printed is an amount in a sentence as the PRD prints one (§8.7): the symbol and digit grouping of the
// currency's home locale, a space only after a letter, no fraction when whole: "£320", "Rp 400.000".
func (c cash) printed(r *big.Rat) string {
	region, _ := language.ParseRegion(c.code[:2])
	base, _ := language.Make("und-" + region.String()).Base()
	tag, _ := language.Compose(base, region)
	p := message.NewPrinter(tag)
	u := currency.MustParseISO(c.code)
	symbol := p.Sprint(currency.Symbol(u))
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
