// Package members is the members service (ARCHITECTURE §7.1): its schema, store and Connect handler.
package members

import (
	"context"
	"embed"
	"errors"
	"log"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1/moneyrecapv1connect"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/migrate"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Migrate applies this service's embedded migrations to the members database (internal/migrate).
func Migrate(ctx context.Context, db *pgxpool.Pool) error { return migrate.Run(ctx, db, migrations) }

// Row is one member row's profile columns; nullable columns read as empty strings.
type Row struct {
	ID, Subject, Email, DisplayName, AvatarKey, CurrencyCode, Timezone string
}

const rowColumns = `id, keycloak_subject, email, coalesce(display_name, ''), coalesce(avatar_key, ''), coalesce(currency_code, ''), coalesce(timezone, '')`

func scanRow(r pgx.Row) (Row, error) {
	var m Row
	err := r.Scan(&m.ID, &m.Subject, &m.Email, &m.DisplayName, &m.AvatarKey, &m.CurrencyCode, &m.Timezone)
	return m, err
}

// Upsert is the one statement behind GetOrCreateBySubject: creates on first sight, refreshes the email
// on every call that carries one (an empty one keeps the stored address), and under two concurrent
// first calls the unique index makes the loser update.
func Upsert(ctx context.Context, db *pgxpool.Pool, subject, email string) (Row, error) {
	return scanRow(db.QueryRow(ctx, `
		INSERT INTO member (keycloak_subject, email) VALUES ($1, $2)
		ON CONFLICT (keycloak_subject) DO UPDATE SET email = coalesce(nullif(EXCLUDED.email, ''), member.email)
		RETURNING `+rowColumns, subject, email))
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
	m, err := Get(ctx, s.DB, req.Msg.MemberId)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && m.Subject != auth.From(ctx).Subject) {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(&moneyrecapv1.GetProfileResponse{Profile: profile(m)}), nil
}

// storeFailed answers the caller's own expired deadline or cancel with its code, unlogged; any other
// store error is logged and answers unavailable, so no driver text (database role and name, an
// address) reaches the caller.
func storeFailed(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return connect.NewError(connect.CodeDeadlineExceeded, ctx.Err())
	}
	if ctx.Err() != nil {
		return connect.NewError(connect.CodeCanceled, ctx.Err())
	}
	log.Printf("members store: %v", err)
	return connect.NewError(connect.CodeUnavailable, errors.New("the members store is unavailable"))
}

func profile(m Row) *moneyrecapv1.Profile {
	return &moneyrecapv1.Profile{MemberId: m.ID, Email: m.Email, DisplayName: m.DisplayName, AvatarKey: m.AvatarKey, CurrencyCode: m.CurrencyCode, Timezone: m.Timezone}
}
