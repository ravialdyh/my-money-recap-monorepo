package ledger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// fixtureDir is packages/contract/fixtures/ledger, relative to this file (ARCHITECTURE §7.2 "Tests").
const fixtureDir = "../../../../packages/contract/fixtures/ledger"

// manifest names every fixture the folder must hold, alphabetically, one line per PRD example; a
// file without a line, or a line without a file, fails TestFixturesMatchManifest.
var manifest = []string{
	"prd-12.4-idr-salary-10th",
	"prd-12.4-idr-salary-25th",
	"prd-12.5-idr-warning-fires",
	"prd-12.5-idr-warning-fires-once",
	"prd-12.6-squirrel-kept-116",
	"prd-8.10-maya-20-september-no-sheet",
	"prd-8.10-maya-27-september-next-month",
	"prd-8.10-maya-27-september-this-month",
	"prd-8.11-signup-27-september",
	"prd-8.2-payday-31-across-months",
	"prd-8.2-payday-last-across-months",
	"prd-8.3-maya-20-september",
	"prd-8.3-maya-30-september",
	"prd-8.4-budi-idr",
	"prd-8.5-signup-4-september-first-run",
	"prd-8.5-signup-4-september-recap",
	"prd-8.6-maya-september-recap",
	"prd-8.7-maya-1-october",
	"prd-8.7-maya-26-october",
	"prd-8.7-maya-300-over",
	"prd-8.7-maya-31-october",
	"prd-8.8-maya-20-september-before-payday",
	"prd-8.8-maya-27-october-after-payday",
	"prd-8.9-maya-gbp-to-idr",
}

// amountPattern is PRD D57: a decimal string with at most 11 digits before the separator.
var amountPattern = regexp.MustCompile(`^\d{1,11}(\.\d+)?$`)

func fixtureNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	slices.Sort(names)
	return names
}

// manifestMismatch says how the folder disagrees with want, or nothing when they match.
func manifestMismatch(t *testing.T, dir string, want []string) string {
	t.Helper()
	if got := fixtureNames(t, dir); !slices.Equal(got, want) {
		return fmt.Sprintf("fixture files %v, manifest %v", got, want)
	}
	return ""
}

func TestFixturesMatchManifest(t *testing.T) {
	if msg := manifestMismatch(t, fixtureDir, manifest); msg != "" {
		t.Fatal(msg)
	}
}

func TestManifestSeesAMissingAndAnExtraFile(t *testing.T) {
	dir := t.TempDir()
	for _, name := range manifest[1:] {
		if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if manifestMismatch(t, dir, manifest) == "" {
		t.Fatal("a missing file passed")
	}
	if err := os.WriteFile(filepath.Join(dir, "prd-0-extra.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if manifestMismatch(t, dir, manifest[1:]) == "" {
		t.Fatal("an extra file passed")
	}
}

func TestFixturesDecode(t *testing.T) {
	for _, name := range fixtureNames(t, fixtureDir) {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(fixtureDir, name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var f moneyrecapv1.LedgerFixture
			if err := protojson.Unmarshal(raw, &f); err != nil {
				t.Fatal(err)
			}
			if f.Doc == "" || f.AsOf == "" {
				t.Fatalf("doc %q and as_of %q must both name what the fixture proves", f.Doc, f.AsOf)
			}
			if !hasExpectation(&f) {
				t.Fatal("no expected value")
			}
			if bad := badAmounts(f.ProtoReflect(), ""); len(bad) != 0 {
				t.Fatalf("amounts breaking D57: %v", bad)
			}
			if bad := badOriginals(&f); len(bad) != 0 {
				t.Fatalf("original amounts breaking the README rule: %v", bad)
			}
		})
	}
}

func TestFixtureRejectsUnknownField(t *testing.T) {
	var f moneyrecapv1.LedgerFixture
	if err := protojson.Unmarshal([]byte(`{"doc":"PRD §8.3","asOf":"2026-09-20","bogus":1}`), &f); err == nil {
		t.Fatal("unknown field decoded")
	}
}

func TestBadAmountsReachesANestedMoney(t *testing.T) {
	f := &moneyrecapv1.LedgerFixture{Spends: []*moneyrecapv1.Spend{{Amount: &moneyrecapv1.Money{Amount: "1,760"}}}}
	if bad := badAmounts(f.ProtoReflect(), ""); len(bad) != 1 || bad[0] != `.spends[0].amount "1,760"` {
		t.Fatalf("badAmounts: %v", bad)
	}
}

func TestBadOriginalsSeesEachDrift(t *testing.T) {
	gbp := &moneyrecapv1.Money{Amount: "2400.00", CurrencyCode: "GBP"}
	idr := &moneyrecapv1.Money{Amount: "48360000", CurrencyCode: "IDR"}
	convert := &moneyrecapv1.ApplyCurrencyConversionRequest{Rate: "20150", ToCurrencyCode: "IDR"}
	for want, f := range map[string]*moneyrecapv1.LedgerFixture{
		".settings.next_salary has no original_amount": {
			Settings: &moneyrecapv1.LedgerSettings{NextSalary: &moneyrecapv1.SalarySetting{Amount: gbp}},
		},
		".recaps[0].splurge original_amount 2400.00 GBP, want 420.00 GBP": {
			Recaps: []*moneyrecapv1.Recap{{Splurge: &moneyrecapv1.Spend{
				Amount: &moneyrecapv1.Money{Amount: "420.00", CurrencyCode: "GBP"}, OriginalAmount: gbp,
			}}},
		},
		".spends[0] original_amount 48360000 IDR, want 2400.00 GBP": {
			Convert: convert,
			Spends:  []*moneyrecapv1.Spend{{Id: "maya-01", Amount: gbp, OriginalAmount: idr}},
		},
		".salaries[0] original_amount 48360000 IDR, want 2400.00 GBP": {
			Convert:  convert,
			Salaries: []*moneyrecapv1.SalarySetting{{Amount: gbp, EffectiveFrom: "2026-08", OriginalAmount: idr}},
		},
		".salaries_after[0] original_amount 48360000 IDR, want 2400.00 GBP": {
			Convert:       convert,
			Salaries:      []*moneyrecapv1.SalarySetting{{Amount: gbp, EffectiveFrom: "2026-08", OriginalAmount: gbp}},
			SalariesAfter: []*moneyrecapv1.SalarySetting{{Amount: idr, EffectiveFrom: "2026-08", OriginalAmount: idr}},
		},
		".salaries_after[0] has no input row salary 2026-09": {
			Convert:       convert,
			Salaries:      []*moneyrecapv1.SalarySetting{{Amount: gbp, EffectiveFrom: "2026-08", OriginalAmount: gbp}},
			SalariesAfter: []*moneyrecapv1.SalarySetting{{Amount: idr, EffectiveFrom: "2026-09", OriginalAmount: gbp}},
		},
	} {
		if bad := badOriginals(f); !slices.Equal(bad, []string{want}) {
			t.Errorf("want [%s], got %v", want, bad)
		}
	}
}

func TestAmountPattern(t *testing.T) {
	for amount, ok := range map[string]bool{"0": true, "2400.00": true, "99999999999": true, "999999999999": false, "1,760": false, "": false, "-1": false} {
		if amountPattern.MatchString(amount) != ok {
			t.Errorf("%q: want %v", amount, ok)
		}
	}
}

func hasExpectation(f *moneyrecapv1.LedgerFixture) bool {
	return len(f.Months)+len(f.Recaps)+len(f.SalariesAfter)+len(f.SpendsAfter) != 0 ||
		f.Settings != nil || f.ConversionPreview != nil || f.CrossedLine != nil
}

// eachMessage calls visit with m and every message populated under it, each by its path; a list
// element's path carries its index.
func eachMessage(m protoreflect.Message, path string, visit func(protoreflect.Message, string)) {
	visit(m, path)
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		child := path + "." + string(fd.Name())
		switch {
		case fd.IsList() && fd.Kind() == protoreflect.MessageKind:
			for i := range v.List().Len() {
				eachMessage(v.List().Get(i).Message(), fmt.Sprintf("%s[%d]", child, i), visit)
			}
		case fd.Kind() == protoreflect.MessageKind:
			eachMessage(v.Message(), child, visit)
		}
		return true
	})
}

// badAmounts walks every populated field and returns each Money whose amount breaks D57, by path.
func badAmounts(m protoreflect.Message, path string) []string {
	var bad []string
	eachMessage(m, path, func(msg protoreflect.Message, at string) {
		if money, ok := msg.Interface().(*moneyrecapv1.Money); ok && !amountPattern.MatchString(money.Amount) {
			bad = append(bad, fmt.Sprintf("%s %q", at, money.Amount))
		}
	})
	return bad
}

// badOriginals returns, by path, each Spend and SalarySetting breaking the README's original_amount
// rule: always present, equal to amount until convert, then equal to the input row's (a Spend by id,
// a SalarySetting by effective_from).
func badOriginals(f *moneyrecapv1.LedgerFixture) []string {
	typed := map[string]*moneyrecapv1.Money{}
	for _, s := range f.Salaries {
		typed["salary "+s.EffectiveFrom] = s.OriginalAmount
	}
	for _, s := range f.Spends {
		typed["spend "+s.Id] = s.OriginalAmount
	}
	var bad []string
	eachMessage(f.ProtoReflect(), "", func(msg protoreflect.Message, at string) {
		var key string
		var amount, original *moneyrecapv1.Money
		var input bool
		switch row := msg.Interface().(type) {
		case *moneyrecapv1.SalarySetting:
			key, amount, original, input = "salary "+row.EffectiveFrom, row.Amount, row.OriginalAmount, slices.Contains(f.Salaries, row)
		case *moneyrecapv1.Spend:
			key, amount, original, input = "spend "+row.Id, row.Amount, row.OriginalAmount, slices.Contains(f.Spends, row)
		default:
			return
		}
		want := amount
		if f.Convert != nil && !input {
			var paired bool
			if want, paired = typed[key]; !paired {
				bad = append(bad, at+" has no input row "+key)
				return
			}
		}
		switch {
		case original == nil:
			bad = append(bad, at+" has no original_amount")
		case !proto.Equal(original, want):
			bad = append(bad, fmt.Sprintf("%s original_amount %s %s, want %s %s",
				at, original.Amount, original.CurrencyCode, want.GetAmount(), want.GetCurrencyCode()))
		}
	})
	return bad
}
