ALTER TABLE member_facts ADD COLUMN facts_at timestamptz;

CREATE TABLE processed_events (
    message_id   text        PRIMARY KEY,
    processed_at timestamptz NOT NULL DEFAULT now()
);
