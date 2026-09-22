CREATE TABLE member_facts (
    member_id        uuid PRIMARY KEY,
    keycloak_subject text,
    timezone         text,
    signed_up_on     date,
    currency_code    char(3)
);

CREATE TABLE salary_setting (
    id                bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    member_id         uuid        NOT NULL,
    amount            numeric     NOT NULL CHECK (amount > 0),
    original_amount   numeric     NOT NULL CHECK (original_amount > 0),
    original_currency char(3)     NOT NULL,
    payday            text        NOT NULL CHECK (payday = 'last' OR payday ~ '^([1-9]|[12][0-9]|3[01])$'),
    effective_from    date        NOT NULL CHECK (extract(day FROM effective_from) = 1),
    created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX salary_setting_member ON salary_setting (member_id, id);

CREATE TABLE saving_target_setting (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    member_id      uuid        NOT NULL,
    percent        smallint    NOT NULL CHECK (percent BETWEEN 0 AND 100),
    effective_from date        NOT NULL CHECK (extract(day FROM effective_from) = 1),
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX saving_target_setting_member ON saving_target_setting (member_id, id);

CREATE TABLE idempotency (
    member_id       uuid        NOT NULL,
    idempotency_key text        NOT NULL,
    request_hash    bytea       NOT NULL,
    response        bytea,
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (member_id, idempotency_key)
);
CREATE INDEX idempotency_created_at ON idempotency (created_at);
