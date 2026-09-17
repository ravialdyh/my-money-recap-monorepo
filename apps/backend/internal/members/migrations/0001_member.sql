CREATE TABLE member (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    keycloak_subject      text NOT NULL UNIQUE,
    email                 text NOT NULL,
    display_name          text,
    avatar_key            text,
    currency_code         char(3),
    timezone              text,
    created_at            timestamptz NOT NULL DEFAULT now(),
    signed_up_on          date,
    deletion_requested_at timestamptz,
    deleted_at            timestamptz,
    onboarding_step       smallint
);
