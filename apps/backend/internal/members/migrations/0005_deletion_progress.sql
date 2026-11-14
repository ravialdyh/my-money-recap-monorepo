CREATE TABLE deletion_progress (
    member_id           uuid        PRIMARY KEY,
    keycloak_subject    text        NOT NULL,
    deleted_at          timestamptz NOT NULL,
    keycloak_deleted_at timestamptz,
    ledger_purged_at    timestamptz,
    notify_purged_at    timestamptz
);
