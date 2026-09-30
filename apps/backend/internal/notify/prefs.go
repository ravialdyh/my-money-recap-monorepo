package notify

import (
	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// defaults is what a Member with no preference row hears: PRD §10's three switches and §6.16's 20:00.
func defaults() *moneyrecapv1.NotificationPrefs {
	return &moneyrecapv1.NotificationPrefs{RecapReady: true, ReminderTime: "20:00", TargetWarning: true}
}

// GetPrefs answers the Member's one row, or the defaults before any save.
func (s *Service) GetPrefs(ctx context.Context, req *connect.Request[moneyrecapv1.GetPrefsRequest]) (*connect.Response[moneyrecapv1.GetPrefsResponse], error) {
	if err := s.owns(ctx, req.Msg.MemberId); err != nil {
		return nil, err
	}
	p := defaults()
	err := s.DB.QueryRow(ctx, `SELECT recap_ready, daily_reminder, to_char(daily_reminder_time, 'HH24:MI'), target_warning
		FROM notification_preference WHERE member_id = $1`, req.Msg.MemberId).Scan(&p.RecapReady, &p.DailyReminder, &p.ReminderTime, &p.TargetWarning)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, storeFailed(ctx, err)
	}
	return connect.NewResponse(&moneyrecapv1.GetPrefsResponse{Prefs: p}), nil
}

// SavePrefs replaces the Member's row with the whole message under their key (A14), answering it as stored. The
// reminder time is "HH:MM" on a 24-hour clock.
func (s *Service) SavePrefs(ctx context.Context, req *connect.Request[moneyrecapv1.SavePrefsRequest]) (*connect.Response[moneyrecapv1.SavePrefsResponse], error) {
	m := req.Msg
	if m.IdempotencyKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("idempotency_key is required"))
	}
	if m.Prefs == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("prefs is required"))
	}
	at, err := time.Parse("15:04", m.Prefs.ReminderTime)
	if err != nil || len(m.Prefs.ReminderTime) != len("15:04") {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("reminder_time is not HH:MM"))
	}
	if err := s.owns(ctx, m.MemberId); err != nil {
		return nil, err
	}
	p := &moneyrecapv1.NotificationPrefs{RecapReady: m.Prefs.RecapReady, DailyReminder: m.Prefs.DailyReminder, ReminderTime: at.Format("15:04"), TargetWarning: m.Prefs.TargetWarning}
	answer := &moneyrecapv1.SavePrefsResponse{Prefs: p}
	err = s.write(ctx, m.MemberId, m.IdempotencyKey, m, answer, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO notification_preference (member_id, recap_ready, daily_reminder, daily_reminder_time, target_warning)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (member_id) DO UPDATE SET recap_ready = EXCLUDED.recap_ready, daily_reminder = EXCLUDED.daily_reminder,
				daily_reminder_time = EXCLUDED.daily_reminder_time, target_warning = EXCLUDED.target_warning`,
			m.MemberId, p.RecapReady, p.DailyReminder, p.ReminderTime, p.TargetWarning)
		return err
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(answer), nil
}
