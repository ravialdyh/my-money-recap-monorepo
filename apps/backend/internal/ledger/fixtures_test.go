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
	if bad := badAmounts(f.ProtoReflect(), ""); len(bad) != 1 || bad[0] != `.spends.amount "1,760"` {
		t.Fatalf("badAmounts: %v", bad)
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

// badAmounts walks every populated field and returns each Money whose amount breaks D57, by path.
func badAmounts(m protoreflect.Message, path string) []string {
	if money, ok := m.Interface().(*moneyrecapv1.Money); ok {
		if amountPattern.MatchString(money.Amount) {
			return nil
		}
		return []string{fmt.Sprintf("%s %q", path, money.Amount)}
	}
	var bad []string
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		child := path + "." + string(fd.Name())
		switch {
		case fd.IsList() && fd.Kind() == protoreflect.MessageKind:
			for i := range v.List().Len() {
				bad = append(bad, badAmounts(v.List().Get(i).Message(), child)...)
			}
		case fd.Kind() == protoreflect.MessageKind:
			bad = append(bad, badAmounts(v.Message(), child)...)
		}
		return true
	})
	return bad
}
