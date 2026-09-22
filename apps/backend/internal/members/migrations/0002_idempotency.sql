CREATE TABLE idempotency (
    member_id       uuid        NOT NULL REFERENCES member (id) ON DELETE CASCADE,
    idempotency_key text        NOT NULL,
    request_hash    bytea       NOT NULL,
    response        bytea,
    created_at      timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (member_id, idempotency_key)
);
CREATE INDEX idempotency_created_at ON idempotency (created_at);
