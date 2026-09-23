package ledger

import (
	"math/big"
	"strings"
	"testing"
)

func rat(s string) *big.Rat {
	r, _ := new(big.Rat).SetString(s)
	return r
}

// Amounts carry the decimals CLDR gives the Currency (package cldr), rounded half up (PRD §8.4), and a sentence
// prints them as the PRD does: "£320", "Rp 400.000", a fraction only when there is one.
func TestAmountsFollowTheCurrencysDecimals(t *testing.T) {
	gbp, err := currencyOf("GBP")
	if err != nil || gbp.scale != 2 {
		t.Fatalf("GBP: %v, %v, want 2 decimals", gbp, err)
	}
	idr, err := currencyOf("IDR")
	if err != nil || idr.scale != 0 {
		t.Fatalf("IDR: %v, %v, want 0 decimals", idr, err)
	}
	for _, c := range []struct{ got, want string }{
		{gbp.money(rat("2400")).Amount, "2400.00"},
		{gbp.money(rat("480.005")).Amount, "480.01"},
		{idr.money(rat("600000.5")).Amount, "600001"},
		{gbp.printed(rat("320")), "£320"},
		{gbp.printed(rat("1920.4")), "£1,920.40"},
		{idr.printed(rat("400000")), "Rp 400.000"},
	} {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
	for _, code := range []string{"", "GB", "ZZZ", "gbp1", "gbp", "XAU"} {
		if _, err := currencyOf(code); err == nil {
			t.Errorf("currencyOf(%q) accepted a code that is no currency", code)
		}
	}
}

// A currency CLDR knows but the older x/text symbols lack prints with its code for a symbol rather than stopping
// the sentence, since step 2 accepts every code CLDR does (§7.2).
func TestASentencePrintsEveryCurrency(t *testing.T) {
	for _, code := range []string{"MRU", "SLE", "VES", "XCG", "ZWG"} {
		c, err := currencyOf(code)
		if err != nil {
			t.Fatalf("currencyOf(%q): %v", code, err)
		}
		if got := c.printed(rat("1500")); !strings.HasPrefix(got, code+" 1") || !strings.HasSuffix(got, "500") {
			t.Errorf("%s printed %q, want its code, a space, then 1500 grouped", code, got)
		}
	}
}
