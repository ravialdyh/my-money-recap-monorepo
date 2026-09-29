package notify

import (
	"context"
	"errors"
	"log"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/auth"
	"github.com/ravialdyh/my-money-recap-monorepo/apps/backend/internal/serve"
)

// GetWarningState answers whether the Member's month is stamped and the copy kept when it was (PRD §8.8).
func (s *Service) GetWarningState(ctx context.Context, req *connect.Request[moneyrecapv1.GetWarningStateRequest]) (*connect.Response[moneyrecapv1.GetWarningStateResponse], error) {
	if _, err := time.Parse("2006-01", req.Msg.Month); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("month is not YYYY-MM"))
	}
	if err := s.owns(ctx, req.Msg.MemberId); err != nil {
		return nil, err
	}
	warning, err := s.kept(ctx, req.Msg.MemberId, req.Msg.Month)
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(&moneyrecapv1.GetWarningStateResponse{Fired: warning != nil, Warning: warning}), nil
}

// EvaluateWarning stamps the month with ledger's copy when crossed_line and unstamped, and sends nothing yet. The
// month flag's primary key is its A14 guarantee: a repeat answers the same kept copy, and only one call sees fired_now.
func (s *Service) EvaluateWarning(ctx context.Context, req *connect.Request[moneyrecapv1.EvaluateWarningRequest]) (*connect.Response[moneyrecapv1.EvaluateWarningResponse], error) {
	m := req.Msg
	if m.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	if _, err := time.Parse("2006-01", m.Month); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("month is not YYYY-MM"))
	}
	if m.CrossedLine && m.Warning.GetMessage() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("crossed_line needs ledger's warning copy"))
	}
	if err := s.owns(ctx, m.MemberId); err != nil {
		return nil, err
	}
	answer := &moneyrecapv1.EvaluateWarningResponse{}
	if m.CrossedLine {
		stamped, err := s.DB.Exec(ctx, `INSERT INTO month_flag (member_id, year_month, target_warning_message) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, m.MemberId, m.Month, m.Warning.Message)
		if err != nil {
			return nil, storeFailed(ctx, err)
		}
		answer.FiredNow = stamped.RowsAffected() == 1
	}
	warning, err := s.kept(ctx, m.MemberId, m.Month)
	if err != nil {
		return nil, storeFailed(ctx, err)
	}
	answer.Fired, answer.Warning = warning != nil, warning
	return connect.NewResponse(answer), nil
}

// Queue is notify's queue of ledger's Spend events, and Keys what it binds (ARCHITECTURE §4.3, §7.3 "Consumes"): a
// delete never fires a warning (PRD §8.8), so ledger.spend.deleted is not bound.
const Queue = "notify.ledger"

var Keys = []string{"ledger.spend.logged", "ledger.spend.edited"}

// StampFromEvent is the event backup (§6.2 step 6) for ledger.spend.logged and ledger.spend.edited (one shape): with
// crossed_line it stamps the event's month with its copy if unstamped, as EvaluateWarning does. It sends no push yet.
func StampFromEvent(ctx context.Context, tx pgx.Tx, d amqp.Delivery) error {
	var e moneyrecapv1.LedgerSpendLogged
	if err := proto.Unmarshal(d.Body, &e); err != nil {
		return err
	}
	if !e.CrossedLine {
		return nil
	}
	id, month, message := e.GetEnvelope().GetMemberId(), e.GetMonth().GetMonth(), e.GetMonth().GetTargetWarning().GetMessage()
	if id == "" || month == "" || message == "" {
		return errors.New("a crossing event names no Member, no month or no warning copy")
	}
	_, err := tx.Exec(ctx, `INSERT INTO month_flag (member_id, year_month, target_warning_message) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, id, month, message)
	return err
}

func (s *Service) kept(ctx context.Context, id, month string) (*moneyrecapv1.TargetWarning, error) {
	var message string
	err := s.DB.QueryRow(ctx, `SELECT target_warning_message FROM month_flag WHERE member_id = $1 AND year_month = $2`, id, month).Scan(&message)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &moneyrecapv1.TargetWarning{Message: message}, nil
}

// owns is ARCHITECTURE §4.2 point 3: a token other than the one the replica kept, or a row lacking a zone or an
// email, is a miss members.GetProfile answers under that token; the read fills both unless a newer event has.
func (s *Service) owns(ctx context.Context, id string) error {
	if err := (&pgtype.UUID{}).Scan(id); err != nil {
		return connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	subject := auth.From(ctx).Subject
	var kept string
	var complete bool
	err := s.DB.QueryRow(ctx, `SELECT coalesce(keycloak_subject, ''), timezone IS NOT NULL AND email IS NOT NULL FROM member_facts WHERE member_id = $1`, id).Scan(&kept, &complete)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return storeFailed(ctx, err)
	}
	if kept != "" && kept == subject && complete {
		return nil
	}
	asked := time.Now()
	res, err := s.Members.GetProfile(ctx, connect.NewRequest(&moneyrecapv1.GetProfileRequest{MemberId: id}))
	if connect.CodeOf(err) == connect.CodeNotFound {
		return connect.NewError(connect.CodeNotFound, errors.New("no such member"))
	}
	if err != nil {
		log.Printf("notify: members.GetProfile: %v", err)
		return connect.NewError(connect.CodeUnavailable, errors.New("members.GetProfile failed"))
	}
	p := res.Msg.GetProfile()
	if _, err := s.DB.Exec(ctx, `
		INSERT INTO member_facts (member_id, keycloak_subject, timezone, email, facts_at) VALUES ($1, $2, nullif($3, ''), nullif($4, ''), $5)
		ON CONFLICT (member_id) DO UPDATE SET keycloak_subject = EXCLUDED.keycloak_subject,
			timezone = CASE WHEN member_facts.facts_at > EXCLUDED.facts_at THEN member_facts.timezone ELSE coalesce(EXCLUDED.timezone, member_facts.timezone) END,
			email = CASE WHEN member_facts.facts_at > EXCLUDED.facts_at THEN member_facts.email ELSE coalesce(EXCLUDED.email, member_facts.email) END,
			facts_at = greatest(member_facts.facts_at, EXCLUDED.facts_at)`,
		id, subject, p.GetTimezone(), p.GetEmail(), asked); err != nil {
		return storeFailed(ctx, err)
	}
	return nil
}

// storeFailed is serve.StoreFailed for this service.
func storeFailed(ctx context.Context, err error) error { return serve.StoreFailed(ctx, "notify", err) }
