package ledger

import (
	"bytes"
	"context"
	"encoding/csv"
	"strings"

	"connectrpc.com/connect"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// ExportCSV is every non-deleted Spend, oldest first, under the header date,category,amount,currency,note (PRD §6.16,
// D39): each amount the exact decimal in the current Currency, quoted by RFC 4180, only the header with no Spends.
func (s *Service) ExportCSV(ctx context.Context, req *connect.Request[moneyrecapv1.ExportCSVRequest]) (*connect.Response[moneyrecapv1.ExportCSVResponse], error) {
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	w := csv.NewWriter(&b)
	w.UseCRLF = true
	rows := [][]string{{"date", "category", "amount", "currency", "note"}}
	for _, sp := range l.spends {
		label := strings.TrimPrefix(sp.spend.Category.String(), "CATEGORY_")
		row := []string{sp.spend.Date, label[:1] + strings.ToLower(label[1:]), sp.spend.Amount.Amount, l.cash.code, sp.spend.Note}
		for i := range row {
			row[i] = guarded(row[i])
		}
		rows = append(rows, row)
	}
	if err := w.WriteAll(rows); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&moneyrecapv1.ExportCSVResponse{Csv: b.Bytes()}), nil
}

// guarded puts an apostrophe before a cell a spreadsheet would run as a formula: one starting with =, +, - or @.
func guarded(cell string) string {
	if cell != "" && strings.ContainsRune("=+-@", rune(cell[0])) {
		return "'" + cell
	}
	return cell
}
