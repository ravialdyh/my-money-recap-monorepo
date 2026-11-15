package ledger

import (
	"errors"
	"math/big"
	"strings"

	"golang.org/x/text/currency"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/cldr"
)

var errNotCurrency = errors.New("currency_code is not an ISO 4217 currency")

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

func (c cash) money(r *big.Rat) *moneyrecapv1.Money {
	return &moneyrecapv1.Money{Amount: r.FloatString(c.scale), CurrencyCode: c.code}
}

func (c cash) rounded(r *big.Rat) *big.Rat {
	v, _ := new(big.Rat).SetString(r.FloatString(c.scale))
	return v
}

func (c cash) unit() *big.Rat {
	return new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(c.scale)), nil))
}

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
	if symbol == c.code {
		return symbol + " " + amount
	}
	return strings.TrimSpace(symbol) + amount
}
