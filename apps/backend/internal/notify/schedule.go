package notify

import (
	"context"
	"log"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/errgroup"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// recapKind and reminderKind are the two schedules' kinds in sent_log, each named as its switch's column: the period
// is the month the Recap is of, and the Member's local date the reminder is for.
const (
	recapKind    = "recap_ready"
	reminderKind = "daily_reminder"
)

// batch is how many due Members one tick of a schedule takes, and senders how many of them it serves at once (§4.5).
const (
	batch   = 500
	senders = 8
)

// nextRecap and nextReminder are a member_facts row f's next-fire times from $2 in its zone (migration 0010): 09:00 on
// the next 1st while Recap ready is on, the chosen time's next coming while the reminder is on, and NULL otherwise.
const (
	nextRecap    = `CASE WHEN coalesce((SELECT recap_ready FROM notification_preference p WHERE p.member_id = f.member_id), true) THEN next_recap(f.timezone, $2) END`
	nextReminder = `(SELECT next_reminder(f.timezone, p.daily_reminder_time, $2) FROM notification_preference p WHERE p.member_id = f.member_id AND p.daily_reminder)`
)

type execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// reschedule sets both of the Member's next-fire times from now: after a zone change or a preferences save.
func reschedule(ctx context.Context, db execer, id string, now time.Time) error {
	_, err := db.Exec(ctx, `UPDATE member_facts f SET next_recap_at = `+nextRecap+`, next_reminder_at = `+nextReminder+` WHERE f.member_id = $1`, id, now)
	return err
}

// Schedule runs both schedules now and then every minute until ctx ends (ARCHITECTURE §4.5, §7.3).
func (s *Service) Schedule(ctx context.Context) {
	every := time.NewTicker(time.Minute)
	defer every.Stop()
	for {
		s.tick(ctx, time.Now())
		select {
		case <-ctx.Done():
			return
		case <-every.C:
		}
	}
}

func (s *Service) tick(ctx context.Context, now time.Time) {
	if err := s.recapReady(ctx, now); err != nil {
		log.Printf("notify: the %s schedule: %v", recapKind, err)
	}
	if err := s.remind(ctx, now); err != nil {
		log.Printf("notify: the %s schedule: %v", reminderKind, err)
	}
}

// due is one Member whose next-fire time has come, and the period that fire is for.
type due struct{ ID, Period string }

// dueBy is the one indexed query of a schedule's tick: the batch of Members whose column is at or before now, soonest
// first, each with period, an expression over the row.
func (s *Service) dueBy(ctx context.Context, column, period string, now time.Time) ([]due, error) {
	rows, err := s.DB.Query(ctx, `SELECT member_id::text, `+period+` FROM member_facts WHERE `+column+` <= $1 ORDER BY `+column+` LIMIT $2`, now, batch)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[due])
}

// each runs fire for each Member of the batch, senders at a time, then advances that Member's column from now with
// next; a fire that fails leaves the column as it is, so the next tick tries again.
func (s *Service) each(ctx context.Context, members []due, column, next string, now time.Time, fire func(due) error) {
	var g errgroup.Group
	g.SetLimit(senders)
	for _, m := range members {
		g.Go(func() error {
			if err := fire(m); err != nil {
				log.Printf("notify: %s for %s: %v", column, m.Period, err)
				return nil
			}
			if _, err := s.DB.Exec(ctx, `UPDATE member_facts f SET `+column+` = `+next+` WHERE f.member_id = $1`, m.ID, now); err != nil {
				log.Printf("notify: advancing %s: %v", column, err)
			}
			return nil
		})
	}
	_ = g.Wait()
}

// recapReady is PRD §10's Recap ready: at 09:00 local on the 1st, a push when ledger's month list, read with notify's
// own key, holds last month, a partial month included.
func (s *Service) recapReady(ctx context.Context, now time.Time) error {
	members, err := s.dueBy(ctx, "next_recap_at", `to_char((next_recap_at AT TIME ZONE timezone) - interval '1 month', 'YYYY-MM')`, now)
	if err != nil {
		return err
	}
	s.each(ctx, members, "next_recap_at", nextRecap, now, func(m due) error {
		res, err := s.Ledger.ListMonthsWithData(ctx, connect.NewRequest(&moneyrecapv1.ListMonthsWithDataRequest{MemberId: m.ID}))
		if code := connect.CodeOf(err); code == connect.CodeNotFound || code == connect.CodeFailedPrecondition {
			return nil
		}
		if err != nil {
			return err
		}
		if !slices.Contains(res.Msg.Months, m.Period) {
			return nil
		}
		month, err := time.Parse("2006-01", m.Period)
		if err != nil {
			return err
		}
		logFailed(recapKind, m.Period, s.push(ctx, m.ID, recapKind, m.Period, "Your "+month.Format("January")+" recap is ready. See how you did."))
		return nil
	})
	return nil
}

// remind is PRD §10's daily reminder: at the Member's chosen time, a push when ledger's spent-today, one call for the
// batch with notify's own key, counts no Spend dated today.
func (s *Service) remind(ctx context.Context, now time.Time) error {
	members, err := s.dueBy(ctx, "next_reminder_at", `to_char(next_reminder_at AT TIME ZONE timezone, 'YYYY-MM-DD')`, now)
	if err != nil || len(members) == 0 {
		return err
	}
	ids := make([]string, len(members))
	for i, m := range members {
		ids[i] = m.ID
	}
	res, err := s.Ledger.GetSpentToday(ctx, connect.NewRequest(&moneyrecapv1.GetSpentTodayRequest{MemberIds: ids}))
	if err != nil {
		return err
	}
	counts := map[string]int32{}
	for _, e := range res.Msg.Entries {
		counts[e.MemberId] = e.SpendCount
	}
	s.each(ctx, members, "next_reminder_at", nextReminder, now, func(m due) error {
		if count, ok := counts[m.ID]; ok && count == 0 {
			logFailed(reminderKind, m.Period, s.push(ctx, m.ID, reminderKind, m.Period, "Log today's spends before you forget."))
		}
		return nil
	})
	return nil
}

// logFailed logs a push that was claimed and then failed; its sent_log row stays, so it is never sent again.
func logFailed(kind, period string, err error) {
	if err != nil {
		log.Printf("notify: %s push %s: %v", kind, period, err)
	}
}
