package cldr

import (
	_ "embed"
	"encoding/json"
	"strconv"
)

//go:embed currencyData.json
var data []byte

var digits = load(data)

func load(raw []byte) map[string]int {
	var file struct {
		Supplemental struct {
			CurrencyData struct {
				Fractions map[string]struct {
					Digits string `json:"_digits"`
				} `json:"fractions"`
				Region map[string][]map[string]struct {
					Tender string `json:"_tender"`
				} `json:"region"`
			} `json:"currencyData"`
		} `json:"supplemental"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		panic(err)
	}
	set := file.Supplemental.CurrencyData
	out := map[string]int{}
	for _, currencies := range set.Region {
		for _, entry := range currencies {
			for code, use := range entry {
				if use.Tender == "false" {
					continue
				}
				fraction, ok := set.Fractions[code]
				if !ok {
					fraction = set.Fractions["DEFAULT"]
				}
				n, err := strconv.Atoi(fraction.Digits)
				if err != nil {
					panic(err)
				}
				out[code] = n
			}
		}
	}
	return out
}

func Digits(code string) (int, bool) {
	n, ok := digits[code]
	return n, ok
}
