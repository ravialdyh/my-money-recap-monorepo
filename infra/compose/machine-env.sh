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
stack=$("$(dirname "$0")/../secrets/keyset.sh" stack)
for key in $stack google_client_id apple_services_id apple_team_id apple_sign_in_key_id; do
  line "$(tr '[:lower:]' '[:upper:]' <<<"$key")" "$key"
done
