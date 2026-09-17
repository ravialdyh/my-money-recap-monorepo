#!/bin/bash
# Runs once, on a fresh volume. One database and one role each, no cross-database grants (ARCHITECTURE A9).
set -euo pipefail
psql -v ON_ERROR_STOP=1 -U "${POSTGRES_USER:-postgres}" \
  -v keycloak_pw="$KEYCLOAK_DB_PASSWORD" -v members_pw="$MEMBERS_DB_PASSWORD" <<'SQL'
CREATE ROLE keycloak LOGIN PASSWORD :'keycloak_pw';
CREATE ROLE members  LOGIN PASSWORD :'members_pw';
CREATE DATABASE keycloak OWNER keycloak;
CREATE DATABASE members  OWNER members;
-- Postgres grants CONNECT to PUBLIC on every new database; take it back so only the owner can connect.
REVOKE CONNECT ON DATABASE keycloak FROM PUBLIC;
REVOKE CONNECT ON DATABASE members  FROM PUBLIC;
SQL
