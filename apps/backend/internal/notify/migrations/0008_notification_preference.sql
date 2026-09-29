CREATE TABLE notification_preference (
    member_id           uuid    PRIMARY KEY,
    recap_ready         boolean NOT NULL,
    daily_reminder      boolean NOT NULL,
    daily_reminder_time time    NOT NULL,
    target_warning      boolean NOT NULL
);
