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
	"prd-8.3-maya-20-september",
	"prd-8.3-maya-30-september",
	"prd-8.4-budi-idr",
	"prd-8.6-maya-september-recap",
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

func TestFixturesMatchManifest(t *testing.T) {
	if got := fixtureNames(t, fixtureDir); !slices.Equal(got, manifest) {
		t.Fatalf("fixture files %v, manifest %v", got, manifest)
	}
}

func TestManifestSeesAMissingAndAnExtraFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "prd-0-extra.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := fixtureNames(t, dir); slices.Equal(got, manifest) {
		t.Fatalf("%v equals the manifest", got)
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
