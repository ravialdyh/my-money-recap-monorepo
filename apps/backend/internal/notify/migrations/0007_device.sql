CREATE TABLE device (
    apns_token    text        PRIMARY KEY CHECK (apns_token ~ '^[0-9a-f]+$'),
    member_id     uuid        NOT NULL,
    registered_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX device_member_id ON device (member_id);
