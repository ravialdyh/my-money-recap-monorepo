package notify

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	moneyrecapv1 "github.com/ravialdyh/my-money-recap-monorepo/apps/backend/gen/money_recap/v1"
)

// Before any save a Member hears PRD §10's defaults; a save replaces the one row with the whole message, a repeated
// key writes once, and a keyless, malformed or someone else's save is refused and changes nothing (A14).
func TestPrefsAnswerTheDefaultsThenTheWholeSavedMessage(t *testing.T) {
	ann, bob := "00000000-0000-4000-8000-00000000e001", "00000000-0000-4000-8000-00000000e002"
	s := warningService(t, ann, map[string]string{"ann": ann, "bob": bob})
	t.Cleanup(func() {
		for _, table := range []string{"notification_preference", "idempotency", "member_facts"} {
			if _, err := s.DB.Exec(context.Background(), `DELETE FROM `+table+` WHERE member_id = ANY($1)`, []string{ann, bob}); err != nil {
				t.Logf("cleanup %s: %v", table, err)
			}
		}
	})
	get := func(id string) *moneyrecapv1.NotificationPrefs {
		t.Helper()
		res, err := s.GetPrefs(as("ann"), connect.NewRequest(&moneyrecapv1.GetPrefsRequest{MemberId: id}))
		if err != nil {
			t.Fatal(err)
		}
		return res.Msg.Prefs
	}
	save := func(subject, key, id string, p *moneyrecapv1.NotificationPrefs) (*moneyrecapv1.NotificationPrefs, error) {
		res, err := s.SavePrefs(as(subject), connect.NewRequest(&moneyrecapv1.SavePrefsRequest{IdempotencyKey: key, MemberId: id, Prefs: p}))
		if err != nil {
			return nil, err
		}
		return res.Msg.Prefs, nil
	}

	if got := get(ann); !proto.Equal(got, defaults()) || !got.RecapReady || got.DailyReminder || got.ReminderTime != "20:00" || !got.TargetWarning {
		t.Fatalf("before any save: %v, want Recap ready on, the reminder off at 20:00 and the warning on", got)
	}
	reminder := &moneyrecapv1.NotificationPrefs{RecapReady: true, DailyReminder: true, ReminderTime: "07:30"}
	for range 2 {
		saved, err := save("ann", "k1", ann, reminder)
		if err != nil || !proto.Equal(saved, reminder) {
			t.Fatalf("saving the reminder at 07:30: %v (%v), want it answered as sent", saved, err)
		}
	}
	if got := get(ann); !proto.Equal(got, reminder) {
		t.Errorf("after the save: %v, want the whole message, the warning off included", got)
	}
	allOff := &moneyrecapv1.NotificationPrefs{ReminderTime: "22:05"}
	if _, err := save("ann", "k2", ann, allOff); err != nil {
		t.Fatal(err)
	}
	if got := get(ann); !proto.Equal(got, allOff) {
		t.Errorf("a second save: %v, want it to replace the row", got)
	}
	for name, call := range map[string]func() error{
		"no key":   func() error { _, err := save("ann", "", ann, reminder); return err },
		"no prefs": func() error { _, err := save("ann", "k9", ann, nil); return err },
		"not HH:MM": func() error {
			_, err := save("ann", "k9", ann, &moneyrecapv1.NotificationPrefs{ReminderTime: "7:30"})
			return err
		},
		"past midnight": func() error {
			_, err := save("ann", "k9", ann, &moneyrecapv1.NotificationPrefs{ReminderTime: "24:00"})
			return err
		},
		"key reused":     func() error { _, err := save("ann", "k1", ann, allOff); return err },
		"another Member": func() error { _, err := save("bob", "k9", ann, reminder); return err },
	} {
		if err := call(); err == nil {
			t.Errorf("%s: saved, want refused", name)
		}
	}
	if got := get(ann); !proto.Equal(got, allOff) {
		t.Errorf("after the refusals: %v, want the last saved message", got)
	}
	if _, err := s.GetPrefs(as("bob"), connect.NewRequest(&moneyrecapv1.GetPrefsRequest{MemberId: ann})); connect.CodeOf(err) != connect.CodeNotFound {
		t.Errorf("bob reading ann's: %v, want not_found", err)
	}
}
