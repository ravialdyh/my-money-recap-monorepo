CREATE TABLE outbox (
    id          uuid        PRIMARY KEY,
    member_id   uuid        NOT NULL,
    routing_key text        NOT NULL,
    payload     bytea       NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    sent_at     timestamptz
);
CREATE INDEX outbox_unsent ON outbox (created_at) WHERE sent_at IS NULL;
