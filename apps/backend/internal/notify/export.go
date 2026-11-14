package notify

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// exportBoundary separates the export email's text part from its attachment; base64 never holds a hyphen.
const exportBoundary = "mmr-export-boundary"

// EmailExport mails the CSV to member_facts' address under the Member's key (A14), claimed with the send: a failed
// send frees it, a retried key sends once. Its hash leaves the CSV out, so a Spend changed in between still replays.
func (s *Service) EmailExport(ctx context.Context, req *connect.Request[moneyrecapv1.EmailExportRequest]) (*connect.Response[moneyrecapv1.EmailExportResponse], error) {
	m := req.Msg
	if m.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	if err := s.owns(ctx, m.MemberId); err != nil {
		return nil, err
	}
	var to string
	if err := s.DB.QueryRow(ctx, `SELECT email FROM member_facts WHERE member_id = $1`, m.MemberId).Scan(&to); err != nil {
		return nil, storeFailed(ctx, err)
	}
	answer := &moneyrecapv1.EmailExportResponse{}
	claim := &moneyrecapv1.EmailExportRequest{IdempotencyKey: m.IdempotencyKey, MemberId: m.MemberId}
	err := s.write(ctx, m.MemberId, m.IdempotencyKey, claim, answer, func(pgx.Tx) error {
		err := s.send(ctx, to, exportMessage(m.Csv))
		if errors.Is(err, errAddressRefused) {
			refused := connect.NewError(connect.CodeFailedPrecondition, errAddressRefused)
			if detail, derr := connect.NewErrorDetail(&moneyrecapv1.EmailBounced{Address: to, BouncedAt: timestamppb.New(time.Now())}); derr == nil {
				refused.AddDetail(detail)
			}
			return refused
		}
		if err != nil {
			log.Printf("notify export: %v", err)
			return connect.NewError(connect.CodeUnavailable, errors.New("the export email was not sent"))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(answer), nil
}

// exportMessage is the export email after From and To: the copy of spec #169 and the CSV attached as
// my-money-recap-spends.csv, base64 in lines of 76 (RFC 2045).
func exportMessage(csv []byte) string {
	var b strings.Builder
	b.WriteString("Subject: Your spends from My Money Recap\r\nMIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/mixed; boundary=" + exportBoundary + "\r\n\r\n")
	b.WriteString("--" + exportBoundary + "\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString("Here is every spend you've logged, as a CSV file.\r\n")
	b.WriteString("--" + exportBoundary + "\r\nContent-Type: text/csv; charset=utf-8; name=my-money-recap-spends.csv\r\n")
	b.WriteString("Content-Disposition: attachment; filename=my-money-recap-spends.csv\r\nContent-Transfer-Encoding: base64\r\n\r\n")
	encoded := base64.StdEncoding.EncodeToString(csv)
	for len(encoded) > 76 {
		b.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	b.WriteString(encoded + "\r\n--" + exportBoundary + "--\r\n")
	return b.String()
}
