CREATE TABLE magic_link_send (
    idempotency_key text PRIMARY KEY,
    email           text NOT NULL,
    sent_at         timestamptz NOT NULL DEFAULT now()
);
