ALTER TABLE member_facts ALTER COLUMN keycloak_subject DROP NOT NULL;
ALTER TABLE member_facts ADD COLUMN timezone text;
ALTER TABLE member_facts ADD COLUMN email text;
ALTER TABLE member_facts ADD COLUMN facts_at timestamptz;
ALTER TABLE member_facts ADD COLUMN deletion_pending boolean NOT NULL DEFAULT false;
ALTER TABLE member_facts ADD COLUMN deletion_at timestamptz;
