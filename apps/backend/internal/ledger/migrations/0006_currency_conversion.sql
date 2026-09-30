CREATE TABLE currency_conversion (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    member_id  uuid        NOT NULL,
    from_code  char(3)     NOT NULL,
    to_code    char(3)     NOT NULL,
    rate       numeric     NOT NULL CHECK (rate > 0),
    source     smallint    NOT NULL CHECK (source IN (1, 2)),
    applied_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX currency_conversion_member ON currency_conversion (member_id);
