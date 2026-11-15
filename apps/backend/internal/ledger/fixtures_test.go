package ledger

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

const fixtureDir = "../../../../packages/contract/fixtures/ledger"

var manifest = []string{
	"bhd-to-idr",
	"budi-idr",
	"idr-salary-10th",
	"idr-salary-25th",
	"idr-warning-fires",
	"idr-warning-fires-once",
	"maya-1-october",
	"maya-20-september",
	"maya-20-september-before-payday",
	"maya-20-september-no-sheet",
	"maya-26-october",
	"maya-27-october-after-payday",
	"maya-27-september-next-month",
	"maya-27-september-this-month",
	"maya-30-september",
	"maya-300-over",
	"maya-31-october",
	"maya-gbp-to-idr",
	"maya-idr-back-to-gbp",
	"maya-september-recap",
	"payday-31-across-months",
	"payday-last-across-months",
	"signup-27-september",
	"signup-4-september-first-run",
	"signup-4-september-recap",
	"squirrel-kept-116",
}

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
	if err := os.WriteFile(filepath.Join(dir, "extra.json"), []byte("{}"), 0o600); err != nil {
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
				t.Fatalf("amounts past 11 digits: %v", bad)
			}
			if bad := badOriginals(&f); len(bad) != 0 {
				t.Fatalf("original amounts breaking the README rule: %v", bad)
			}
		})
	}
}

func TestFixtureRejectsUnknownField(t *testing.T) {
	var f moneyrecapv1.LedgerFixture
	if err := protojson.Unmarshal([]byte(`{"doc":"x","asOf":"2026-09-20","bogus":1}`), &f); err == nil {
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
		".spends[0] original_amount 6000000 IDR, want 6045000 IDR": {
			CurrencyCode: "IDR",
			Spends:       []*moneyrecapv1.Spend{{Id: "maya-01", Amount: &moneyrecapv1.Money{Amount: "6045000", CurrencyCode: "IDR"}, OriginalAmount: &moneyrecapv1.Money{Amount: "6000000", CurrencyCode: "IDR"}}},
		},
	} {
		if bad := badOriginals(f); !slices.Equal(bad, []string{want}) {
			t.Errorf("want [%s], got %v", want, bad)
		}
	}
}

func TestBadOriginalsTakesAnEarlierConversion(t *testing.T) {
	idr := &moneyrecapv1.Money{Amount: "48360000", CurrencyCode: "IDR"}
	gbp := &moneyrecapv1.Money{Amount: "2400.00", CurrencyCode: "GBP"}
	f := &moneyrecapv1.LedgerFixture{
		CurrencyCode:  "IDR",
		Convert:       &moneyrecapv1.ApplyCurrencyConversionRequest{Rate: "0.0000496", ToCurrencyCode: "GBP"},
		Salaries:      []*moneyrecapv1.SalarySetting{{Amount: idr, EffectiveFrom: "2026-08", OriginalAmount: gbp}},
		SalariesAfter: []*moneyrecapv1.SalarySetting{{Amount: &moneyrecapv1.Money{Amount: "2398.66", CurrencyCode: "GBP"}, EffectiveFrom: "2026-08", OriginalAmount: gbp}},
	}
	if bad := badOriginals(f); len(bad) != 0 {
		t.Errorf("a second conversion's rows: %v, want none", bad)
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

func badAmounts(m protoreflect.Message, path string) []string {
	var bad []string
	eachMessage(m, path, func(msg protoreflect.Message, at string) {
		if money, ok := msg.Interface().(*moneyrecapv1.Money); ok && !amountPattern.MatchString(money.Amount) {
			bad = append(bad, fmt.Sprintf("%s %q", at, money.Amount))
		}
	})
	return bad
}

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
		if input && amount.GetCurrencyCode() == f.CurrencyCode && original.GetCurrencyCode() != f.CurrencyCode {
			want = original
		}
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
