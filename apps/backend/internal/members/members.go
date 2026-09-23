// Package members is the members service (ARCHITECTURE §7.1): its schema, store and Connect handler.
package members

import (
	"context"
	"embed"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/cldr"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/migrate"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies this service's embedded migrations to the members database (internal/migrate).
func Migrate(ctx context.Context, db *pgxpool.Pool) error { return migrate.Run(ctx, db, migrations) }

// Row is one member row's profile columns: a null text or date column reads as an empty string, a null
// onboarding step as 0 and a null deletion stamp as nil, so no unset column is an error.
type Row struct {
	ID, Subject, Email, DisplayName, AvatarKey, CurrencyCode, Timezone, SignedUpOn string
	OnboardingStep                                                                 int32
	DeletionRequestedAt                                                            *time.Time
}

const rowColumns = `id, keycloak_subject, email, coalesce(display_name, ''), coalesce(avatar_key, ''), coalesce(currency_code, ''),
	coalesce(timezone, ''), coalesce(to_char(signed_up_on, 'YYYY-MM-DD'), ''), coalesce(onboarding_step, 0), deletion_requested_at`

func scanRow(r pgx.Row) (Row, error) {
	var m Row
	err := r.Scan(&m.ID, &m.Subject, &m.Email, &m.DisplayName, &m.AvatarKey, &m.CurrencyCode, &m.Timezone, &m.SignedUpOn, &m.OnboardingStep, &m.DeletionRequestedAt)
	return m, err
}

// Upsert backs GetOrCreateBySubject: a first call creates the Member (members.created); a later one stores a
// new email claim (members.updated), and an empty claim keeps the address. Concurrent first calls meet on the
// unique index.
func Upsert(ctx context.Context, db *pgxpool.Pool, subject, email string) (Row, error) {
	var m Row
	err := pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
		created, err := scanRow(tx.QueryRow(ctx, `INSERT INTO member (keycloak_subject, email) VALUES ($1, $2) ON CONFLICT (keycloak_subject) DO NOTHING RETURNING `+rowColumns, subject, email))
		if err == nil {
			m = created
			return stage(ctx, tx, "members.created", created)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		before, err := scanRow(tx.QueryRow(ctx, `SELECT `+rowColumns+` FROM member WHERE keycloak_subject = $1 FOR UPDATE`, subject))
		m = before
		if err != nil || email == "" || email == before.Email {
			return err
		}
		if m, err = scanRow(tx.QueryRow(ctx, `UPDATE member SET email = $2 WHERE id = $1 RETURNING `+rowColumns, before.ID, email)); err != nil {
			return err
		}
		return announce(ctx, tx, before, m)
	})
	return m, err
}

// Get returns pgx.ErrNoRows for an unknown or malformed id.
func Get(ctx context.Context, db *pgxpool.Pool, id string) (Row, error) {
	if err := (&pgtype.UUID{}).Scan(id); err != nil {
		return Row{}, pgx.ErrNoRows
	}
	return scanRow(db.QueryRow(ctx, `SELECT `+rowColumns+` FROM member WHERE id = $1`, id))
}

// Service implements MembersService over a pool. Every RPC runs behind auth.Interceptor, and a
// caller only ever reaches the Member its own token names (§4.2 point 3).
type Service struct {
	moneyrecapv1connect.UnimplementedMembersServiceHandler
	DB *pgxpool.Pool
}

func (s *Service) GetOrCreateBySubject(ctx context.Context, req *connect.Request[moneyrecapv1.GetOrCreateBySubjectRequest]) (*connect.Response[moneyrecapv1.GetOrCreateBySubjectResponse], error) {
	claims := auth.From(ctx)
	if req.Msg.Subject != claims.Subject {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("subject does not match the token"))
	}
	if req.Msg.Email != claims.Email {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("email does not match the token"))
	}
	m, err := Upsert(ctx, s.DB, claims.Subject, claims.Email)
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(&moneyrecapv1.GetOrCreateBySubjectResponse{Profile: profile(m)}), nil
}

func (s *Service) GetProfile(ctx context.Context, req *connect.Request[moneyrecapv1.GetProfileRequest]) (*connect.Response[moneyrecapv1.GetProfileResponse], error) {
	m, err := s.own(ctx, req.Msg.MemberId)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&moneyrecapv1.GetProfileResponse{Profile: profile(m)}), nil
}

// own reads the Member id names, answering not_found for an unknown or malformed id and for one another
// token owns, so no caller learns whether someone else's Member exists.
func (s *Service) own(ctx context.Context, id string) (Row, error) {
	m, err := Get(ctx, s.DB, id)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && m.Subject != auth.From(ctx).Subject) {
		return Row{}, connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	if err != nil {
		return Row{}, storeFailed(ctx, err)
	}
	return m, nil
}

// storeFailed is serve.StoreFailed for this service.
func storeFailed(ctx context.Context, err error) error { return serve.StoreFailed(ctx, "members", err) }

// profile is a stored row as the contract carries it, with the Currency's decimals from the CLDR data ledger rounds
// with, which step 3's amount field allows (spec #92).
func profile(m Row) *moneyrecapv1.Profile {
	digits, _ := cldr.Digits(m.CurrencyCode)
	p := &moneyrecapv1.Profile{
		MemberId: m.ID, Email: m.Email, DisplayName: m.DisplayName, AvatarKey: m.AvatarKey, CurrencyCode: m.CurrencyCode,
		Timezone: m.Timezone, OnboardingStep: m.OnboardingStep, SignedUpOn: m.SignedUpOn, CurrencyDigits: int32(digits),
	}
	if m.DeletionRequestedAt != nil {
		p.DeletionRequestedAt = timestamppb.New(*m.DeletionRequestedAt)
	}
	return p
}
