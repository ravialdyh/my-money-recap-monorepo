#!/usr/bin/env bash
set -euo pipefail

[ $# -eq 1 ] || { echo "usage: $(basename "$0") <decrypted secrets directory>" >&2; exit 2; }
dir=$1

line() {
  local value
  value=$(cat "$dir/$2")
  case $value in
    *"'"* | *$'\n'*)
      echo "$2 holds a quote or a line break, which a single-quoted .env value cannot carry" >&2
      exit 1
      ;;
  esac
  printf "%s='%s'\n" "$1" "$value"
}

line MMR_HOST domain
echo REALM_SSL_REQUIRED=external
echo NGINX_HOST_PORT=80
for key in postgres_password keycloak_db_password members_db_password members_database_url \
  notify_db_password notify_database_url kc_bootstrap_admin_password bff_client_secret; do
  line "$(tr '[:lower:]' '[:upper:]' <<<"$key")" "$key"
done
