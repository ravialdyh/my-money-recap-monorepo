package answers

import (
	"errors"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/cldr"
)

var AvatarKeys = []string{"fox", "otter", "panda", "frog", "owl", "rabbit"}

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

func AvatarKey(k string) error {
	if !slices.Contains(AvatarKeys, k) {
		return errors.New("avatar_key is not one of the six Avatars")
	}
	return nil
}

func CurrencyCode(c string) error {
	if _, ok := cldr.Digits(c); !ok {
		return errors.New("currency_code is not an ISO 4217 currency")
	}
	return nil
}

func OnboardingStep(n int32) error {
	if n < 1 || n > 4 {
		return errors.New("onboarding_step is 1 to 4")
	}
	return nil
}

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
