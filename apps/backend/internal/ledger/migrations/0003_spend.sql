CREATE TABLE spend (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id         uuid        NOT NULL,
    amount            numeric     NOT NULL CHECK (amount > 0),
    original_amount   numeric     NOT NULL CHECK (original_amount > 0),
    original_currency char(3)     NOT NULL,
    category          smallint    NOT NULL CHECK (category BETWEEN 1 AND 8),
    spent_on          date        NOT NULL,
    note              text        NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz
);
CREATE INDEX spend_member_date ON spend (member_id, spent_on) WHERE deleted_at IS NULL;
