package cldr

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// The file is cldr-json 48.2.0's, byte for byte: a formatter or a hand edit fails here, and a new release is a
// deliberate bump that changes this checksum with it (ARCHITECTURE §7.2).
func TestTheFileIsCLDR48Unedited(t *testing.T) {
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != "e50315091b1f34428480f82970be64159ad207af0bdcebe7faa6260d3e34c8a6" {
		t.Fatalf("currencyData.json sha256 %s is not cldr-json 48.2.0's", got)
	}
}

// Decimals are CLDR's own (IDR, HUF and COP 0 since CLDR 48), a currency since withdrawn stays one so a stored
// Member keeps working, and a code no territory takes as tender is refused.
func TestDigits(t *testing.T) {
	for code, want := range map[string]int{
		"GBP": 2, "IDR": 0, "HUF": 0, "COP": 0, "JPY": 0, "BHD": 3, "USD": 2, "EUR": 2,
		"VES": 2, "MRU": 2, "SLE": 2, "XCG": 2, "ZWG": 2,
		"HRK": 2, "BGN": 2, "SLL": 0,
	} {
		if got, ok := Digits(code); !ok || got != want {
			t.Errorf("Digits(%q) = %d, %v, want %d, true", code, got, ok, want)
		}
	}
	for _, code := range []string{"", "gbp", "GB", "GBPX", "ZZZ", "XAU", "XDR", "XTS", "XXX", "CHE", "USN"} {
		if n, ok := Digits(code); ok {
			t.Errorf("Digits(%q) = %d, true, want a refusal", code, n)
		}
	}
}

func TestLoadRefusesAFileItCannotRead(t *testing.T) {
	for _, raw := range []string{"not json", `{"supplemental":{"currencyData":{"fractions":{},"region":{"GB":[{"GBP":{}}]}}}}`} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("load(%q) did not panic", raw)
				}
			}()
			load([]byte(raw))
		}()
	}
}
