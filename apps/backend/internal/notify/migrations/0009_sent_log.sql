CREATE TABLE sent_log (
    member_id uuid        NOT NULL,
    kind      text        NOT NULL,
    period    text        NOT NULL,
    sent_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (member_id, kind, period)
);
