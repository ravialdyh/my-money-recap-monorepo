// Package answers holds the rules a Member's step 1 and step 2 answers meet (PRD §6.4, §6.5), which the BFF
// checks before any service call and members checks again as the data's owner.
package answers

import (
	"errors"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/currency"
	"golang.org/x/text/unicode/norm"
)

// AvatarKeys are the six Avatars, the list Profile.avatar_key's documentation holds (PRD D36).
var AvatarKeys = []string{"fox", "otter", "panda", "frog", "owl", "rabbit"}

// DisplayName is the name as stored: NFC, trimmed, 1 to 30 code points of letters in any script with their
// marks, spaces, hyphens and apostrophes (PRD §6.4). The app counts the same way, so it never accepts a name
// this refuses (spec #92).
func DisplayName(s string) (string, error) {
	name := strings.TrimSpace(norm.NFC.String(s))
	if n := utf8.RuneCountInString(name); n < 1 || n > 30 {
		return "", errors.New("display_name is 1 to 30 characters")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsMark(r) && !strings.ContainsRune(" -'’", r) {
			return "", errors.New("display_name holds only letters, spaces, hyphens and apostrophes")
		}
	}
	return name, nil
}

// AvatarKey refuses a key that is not one of the six.
func AvatarKey(k string) error {
	if !slices.Contains(AvatarKeys, k) {
		return errors.New("avatar_key is not one of the six Avatars")
	}
	return nil
}

// CurrencyCode refuses a code CLDR does not know as a currency, the data ledger rounds with (§7.2).
func CurrencyCode(c string) error {
	if u, err := currency.ParseISO(c); err != nil || u.String() != c {
		return errors.New("currency_code is not an ISO 4217 currency")
	}
	return nil
}

// OnboardingStep refuses a step outside 1 to 4 (PRD §5 step 6).
func OnboardingStep(n int32) error {
	if n < 1 || n > 4 {
		return errors.New("onboarding_step is 1 to 4")
	}
	return nil
}

// Check applies these rules to whichever answers are present and answers the Display Name as it is stored, or
// nil when none was sent. The BFF runs it before any service call, and members again as the data's owner.
func Check(name, avatar, code *string, step *int32) (*string, error) {
	if avatar != nil {
		if err := AvatarKey(*avatar); err != nil {
			return nil, err
		}
	}
	if code != nil {
		if err := CurrencyCode(*code); err != nil {
			return nil, err
		}
	}
	if step != nil {
		if err := OnboardingStep(*step); err != nil {
			return nil, err
		}
	}
	if name == nil {
		return nil, nil
	}
	stored, err := DisplayName(*name)
	return &stored, err
}
