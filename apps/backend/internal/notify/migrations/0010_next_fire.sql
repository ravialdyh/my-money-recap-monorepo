CREATE FUNCTION next_recap(zone text, t timestamptz) RETURNS timestamptz LANGUAGE sql STABLE
    RETURN (date_trunc('month', (t AT TIME ZONE zone) - interval '9 hours') + interval '1 month 9 hours') AT TIME ZONE zone;
CREATE FUNCTION next_reminder(zone text, hhmm time, t timestamptz) RETURNS timestamptz LANGUAGE sql STABLE
    RETURN (((t AT TIME ZONE zone) - hhmm)::date + 1 + hhmm) AT TIME ZONE zone;
ALTER TABLE member_facts ADD COLUMN next_recap_at timestamptz, ADD COLUMN next_reminder_at timestamptz;
CREATE INDEX member_facts_next_recap_at ON member_facts (next_recap_at);
CREATE INDEX member_facts_next_reminder_at ON member_facts (next_reminder_at);
UPDATE member_facts f SET next_recap_at = next_recap(f.timezone, now())
    WHERE coalesce((SELECT recap_ready FROM notification_preference p WHERE p.member_id = f.member_id), true);
UPDATE member_facts f SET next_reminder_at = next_reminder(f.timezone, p.daily_reminder_time, now())
    FROM notification_preference p WHERE p.member_id = f.member_id AND p.daily_reminder;
