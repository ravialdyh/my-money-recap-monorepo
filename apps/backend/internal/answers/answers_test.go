package answers

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// A name in any script passes, stored in NFC and trimmed; digits, emoji, other punctuation and a 31st code point
// do not (PRD §6.4).
func TestDisplayName(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"Maya", "Maya"},
		{"  Budi Santoso ", "Budi Santoso"},
		{"O'Neil", "O'Neil"},
		{"D’Arcy", "D’Arcy"},
		{"Jean-Luc", "Jean-Luc"},
		{"Jose\u0301", "Jos\u00e9"},
		{"हिन्दी", "हिन्दी"},
		{"山田太郎", "山田太郎"},
		{strings.Repeat("a", 30), strings.Repeat("a", 30)},
		{" " + strings.Repeat("b", 30), strings.Repeat("b", 30)},
	} {
		if got, err := DisplayName(c.in); err != nil || got != c.want {
			t.Errorf("DisplayName(%q) = %q, %v; want %q", c.in, got, err, c.want)
		}
	}
	for _, in := range []string{"", "   ", "Maya2", "Maya 😀", "Maya!", "Maya_B", "Maya.", strings.Repeat("a", 31)} {
		if got, err := DisplayName(in); err == nil {
			t.Errorf("DisplayName(%q) = %q, want an error", in, got)
		}
	}
}

// The contract's documentation is the authoritative list (spec #92): AvatarKeys must be exactly the keys the
// Profile.avatar_key comment names, in its order.
func TestAvatarKeysAreTheContractsList(t *testing.T) {
	proto, err := os.ReadFile("../../../../packages/contract/money_recap/v1/members.proto")
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`authoritative list: ([a-z, ]+) \(`).FindSubmatch(proto)
	if m == nil {
		t.Fatal("members.proto names no authoritative Avatar list")
	}
	if listed := strings.Split(string(m[1]), ", "); !slices.Equal(listed, AvatarKeys) {
		t.Fatalf("members.proto lists %v, AvatarKeys is %v", listed, AvatarKeys)
	}
	for _, k := range AvatarKeys {
		if AvatarKey(k) != nil {
			t.Errorf("AvatarKey(%q) refused a listed key", k)
		}
	}
	for _, k := range []string{"", "Fox", "cat", "fox "} {
		if AvatarKey(k) == nil {
			t.Errorf("AvatarKey(%q) accepted", k)
		}
	}
}

func TestCurrencyCodeAndStep(t *testing.T) {
	for _, c := range []string{"GBP", "IDR", "USD", "JPY"} {
		if CurrencyCode(c) != nil {
			t.Errorf("CurrencyCode(%q) refused", c)
		}
	}
	for _, c := range []string{"", "gbp", "GB", "ZZZ", "GBPX"} {
		if CurrencyCode(c) == nil {
			t.Errorf("CurrencyCode(%q) accepted", c)
		}
	}
	for n := int32(0); n <= 5; n++ {
		if err := OnboardingStep(n); (err == nil) != (n >= 1 && n <= 4) {
			t.Errorf("OnboardingStep(%d) = %v", n, err)
		}
	}
}
