package ledger

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// fixtureDir is packages/contract/fixtures/ledger, relative to this file (ARCHITECTURE §7.2 "Tests").
const fixtureDir = "../../../../packages/contract/fixtures/ledger"

// manifest names every fixture the folder must hold, one line per PRD example; a file
// without a line, or a line without a file, fails TestFixturesMatchManifest.
var manifest = []string{
	"prd-8.3-maya-20-september",
	"prd-8.3-maya-30-september",
	"prd-8.4-budi-idr",
	"prd-8.6-maya-september-recap",
}

// amountPattern is PRD D57: a decimal string with at most 11 digits before the separator.
var amountPattern = regexp.MustCompile(`^\d{1,11}(\.\d+)?$`)

// strict rejects a field the contract does not know, so a fixture cannot drift from it.
var strict = protojson.UnmarshalOptions{DiscardUnknown: false}

func fixtureNames(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	sort.Strings(names)
	return names
}

func TestFixturesMatchManifest(t *testing.T) {
	want := append([]string(nil), manifest...)
	sort.Strings(want)
	got := fixtureNames(t)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("fixture files:\n%s\nmanifest:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestFixturesDecode(t *testing.T) {
	for _, name := range fixtureNames(t) {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(fixtureDir, name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			var f moneyrecapv1.LedgerFixture
			if err := strict.Unmarshal(raw, &f); err != nil {
				t.Fatal(err)
			}
			if f.Doc == "" || f.AsOf == "" {
				t.Fatalf("doc %q and as_of %q must both name what the fixture proves", f.Doc, f.AsOf)
			}
			if len(f.Months)+len(f.Recaps)+len(f.SalariesAfter)+len(f.SpendsAfter) == 0 && f.Settings == nil && f.ConversionPreview == nil && f.CrossedLine == nil {
				t.Fatal("no expected value")
			}
			checkAmounts(t, f.ProtoReflect(), "")
		})
	}
}

func TestFixtureRejectsUnknownField(t *testing.T) {
	var f moneyrecapv1.LedgerFixture
	if err := strict.Unmarshal([]byte(`{"doc":"PRD §8.3","asOf":"2026-09-20","bogus":1}`), &f); err == nil {
		t.Fatal("unknown field decoded")
	}
}

func TestAmountPattern(t *testing.T) {
	for amount, ok := range map[string]bool{"0": true, "2400.00": true, "99999999999": true, "999999999999": false, "1,760": false, "": false, "-1": false} {
		if amountPattern.MatchString(amount) != ok {
			t.Errorf("%q: want %v", amount, ok)
		}
	}
}

// checkAmounts walks every populated field and asserts D57 on each Money it finds.
func checkAmounts(t *testing.T, m protoreflect.Message, path string) {
	t.Helper()
	if money, ok := m.Interface().(*moneyrecapv1.Money); ok {
		if !amountPattern.MatchString(money.Amount) {
			t.Errorf("%s: amount %q breaks D57", path, money.Amount)
		}
		return
	}
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		child := path + "." + string(fd.Name())
		switch {
		case fd.IsList() && fd.Kind() == protoreflect.MessageKind:
			for i := range v.List().Len() {
				checkAmounts(t, v.List().Get(i).Message(), child)
			}
		case fd.Kind() == protoreflect.MessageKind && !fd.IsMap():
			checkAmounts(t, v.Message(), child)
		}
		return true
	})
}
