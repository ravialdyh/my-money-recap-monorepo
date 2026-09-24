CREATE TABLE member_facts (
    member_id        uuid PRIMARY KEY,
    keycloak_subject text NOT NULL
);

CREATE TABLE month_flag (
    member_id              uuid        NOT NULL,
    year_month             text        NOT NULL CHECK (year_month ~ '^[0-9]{4}-(0[1-9]|1[0-2])$'),
    target_warning_sent_at timestamptz NOT NULL DEFAULT now(),
    target_warning_message text        NOT NULL CHECK (target_warning_message <> ''),
    PRIMARY KEY (member_id, year_month)
);
