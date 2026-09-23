// Package ledger is the money service (ARCHITECTURE §7.2): the Salary, Saving Target and Spend rows, the
// settings read and every Home figure of PRD §8 over them.
package ledger

import (
	"context"
	"embed"
	"errors"
	"log"
	"math/big"
	"time"
	_ "time/tzdata"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/migrate"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies this service's embedded migrations to the ledger database (internal/migrate).
func Migrate(ctx context.Context, db *pgxpool.Pool) error { return migrate.Run(ctx, db, migrations) }

// Service implements LedgerService over a pool. Members answers the profile read that fills the replica
// on a miss, under the caller's forwarded token; Now is the clock, nil meaning time.Now.
type Service struct {
	moneyrecapv1connect.UnimplementedLedgerServiceHandler
	DB      *pgxpool.Pool
	Members moneyrecapv1connect.MembersServiceClient
	Now     func() time.Time
}

// errNoFacts is a Member members knows too little about for any figure: no zone, no sign-up date or no Currency.
var errNoFacts = errors.New("the Member has no time zone, sign-up date or currency yet")

// member is one Member's ledger as of their local today.
type member struct {
	ledgerOf
	id    string
	today time.Time
}

// load reads the Member's replica and rows. The replica keeps the subject whose token filled it, so another
// token's call is a miss members answers under that token: not_found for a Member it does not own (§4.2).
func (s *Service) load(ctx context.Context, id string) (member, error) {
	if err := (&pgtype.UUID{}).Scan(id); err != nil {
		return member{}, connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	var subject string
	var f facts
	err := s.DB.QueryRow(ctx, `SELECT coalesce(keycloak_subject, ''), coalesce(timezone, ''), signed_up_on, coalesce(currency_code, '') FROM member_facts WHERE member_id = $1`, id).Scan(&subject, &f.zone, &f.signedUp, &f.code)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return member{}, storeFailed(ctx, err)
	}
	if subject != auth.From(ctx).Subject || !f.complete() {
		if f, err = s.fill(ctx, id); err != nil {
			return member{}, err
		}
	}
	loc, err := time.LoadLocation(f.zone)
	if err != nil {
		return member{}, connect.NewError(connect.CodeFailedPrecondition, errNoFacts)
	}
	c, err := currencyOf(f.code)
	if err != nil {
		return member{}, connect.NewError(connect.CodeFailedPrecondition, errNoFacts)
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	y, mo, d := now().In(loc).Date()
	m := member{ledgerOf: ledgerOf{cash: c, signedUp: f.signedUp.Time}, id: id, today: time.Date(y, mo, d, 0, 0, 0, 0, time.UTC)}
	if m.salaries, err = salaries(ctx, s.DB, id); err != nil {
		return member{}, storeFailed(ctx, err)
	}
	if m.targets, err = targets(ctx, s.DB, id); err != nil {
		return member{}, storeFailed(ctx, err)
	}
	if m.spends, err = spends(ctx, s.DB, id); err != nil {
		return member{}, storeFailed(ctx, err)
	}
	return m, nil
}

// facts are the replica's three facts about a Member (ARCHITECTURE §7.2): the zone every "today" is read in,
// the sign-up date that bounds the Streak window and names the sign-up month, and the Currency.
type facts struct {
	zone, code string
	signedUp   pgtype.Date
}

func (f facts) complete() bool { return f.zone != "" && f.signedUp.Valid && f.code != "" }

// fill is the profile-on-a-miss read (§6.2 step 3): members.GetProfile under the caller's token, then the row
// upserted with that token's subject. It fills only facts the row lacks, so a read racing a newer event cannot
// overwrite it.
func (s *Service) fill(ctx context.Context, id string) (facts, error) {
	res, err := s.Members.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: id}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		return facts{}, connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	if err != nil {
		log.Printf("ledger: members.GetProfile: %v", err)
		return facts{}, connect.NewError(connect.CodeUnavailable, errors.New("members.GetProfile failed"))
	}
	p := res.Msg.Profile
	f := facts{zone: p.Timezone, code: p.CurrencyCode}
	if p.SignedUpOn != "" && f.signedUp.Scan(p.SignedUpOn) != nil {
		f.signedUp = pgtype.Date{}
	}
	if !f.complete() {
		return facts{}, connect.NewError(connect.CodeFailedPrecondition, errNoFacts)
	}
	err = s.DB.QueryRow(ctx, `
		INSERT INTO member_facts (member_id, keycloak_subject, timezone, signed_up_on, currency_code) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (member_id) DO UPDATE SET keycloak_subject = EXCLUDED.keycloak_subject,
			timezone = coalesce(member_facts.timezone, EXCLUDED.timezone), signed_up_on = coalesce(member_facts.signed_up_on, EXCLUDED.signed_up_on),
			currency_code = coalesce(member_facts.currency_code, EXCLUDED.currency_code)
		RETURNING timezone, signed_up_on, currency_code`,
		id, auth.From(ctx).Subject, f.zone, f.signedUp, f.code).Scan(&f.zone, &f.signedUp, &f.code)
	if err != nil {
		return facts{}, storeFailed(ctx, err)
	}
	return f, nil
}

// querier is a pool or a transaction, so a write can read back the rows it changed.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func salaries(ctx context.Context, db querier, id string) ([]salaryRow, error) {
	rows, err := db.Query(ctx, `SELECT amount::text, original_amount::text, original_currency, payday, effective_from FROM salary_setting WHERE member_id = $1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (salaryRow, error) {
		var amount, original string
		var row salaryRow
		var from time.Time
		if err := r.Scan(&amount, &original, &row.originalCurrency, &row.payday, &from); err != nil {
			return row, err
		}
		row.amount, _ = new(big.Rat).SetString(amount)
		row.original, _ = new(big.Rat).SetString(original)
		row.from = monthOf(from)
		return row, nil
	})
}

// spends reads every Spend of the Member that is not deleted (PRD §7).
func spends(ctx context.Context, db querier, id string) ([]spendRow, error) {
	rows, err := db.Query(ctx, `SELECT amount::text, spent_on FROM spend WHERE member_id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (spendRow, error) {
		var amount string
		var row spendRow
		err := r.Scan(&amount, &row.date)
		row.amount, _ = new(big.Rat).SetString(amount)
		return row, err
	})
}

func targets(ctx context.Context, db *pgxpool.Pool, id string) ([]targetRow, error) {
	rows, err := db.Query(ctx, `SELECT percent, effective_from FROM saving_target_setting WHERE member_id = $1 ORDER BY id`, id)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (targetRow, error) {
		var row targetRow
		var from time.Time
		var percent int16
		err := r.Scan(&percent, &from)
		row.percent, row.from = int32(percent), monthOf(from)
		return row, err
	})
}

func (s *Service) GetMonthSummary(ctx context.Context, req *connect.Request[moneyrecapv1.GetMonthSummaryRequest]) (*connect.Response[moneyrecapv1.GetMonthSummaryResponse], error) {
	m, err := parseMonth(req.Msg.Month)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.GetMonthSummaryResponse{Summary: l.summary(m, l.today)}), nil
}

func (s *Service) GetSettings(ctx context.Context, req *connect.Request[moneyrecapv1.GetSettingsRequest]) (*connect.Response[moneyrecapv1.GetSettingsResponse], error) {
	l, err := s.load(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.GetSettingsResponse{Settings: l.settings(l.today)}), nil
}

// storeFailed is serve.StoreFailed for this service.
func storeFailed(ctx context.Context, err error) error { return serve.StoreFailed(ctx, "ledger", err) }
