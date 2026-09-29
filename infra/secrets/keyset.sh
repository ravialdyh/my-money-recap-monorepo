#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $(basename "$0") keys <example.yaml>" >&2
  echo "       $(basename "$0") value <example.yaml> <key>" >&2
  echo "       $(basename "$0") stack" >&2
  exit 2
}

case ${1:-} in
  keys)
    [ $# -eq 2 ] || usage
    sed -n 's/^\([a-z0-9_]*\):.*/\1/p' "$2"
    ;;
  value)
    [ $# -eq 3 ] || usage
    sed -n "s/^$3: //p" "$2"
    ;;
  stack)
    [ $# -eq 1 ] || usage
    printf '%s\n' members_db_password members_database_url notify_db_password notify_database_url \
      ledger_db_password ledger_database_url rabbitmq_password amqp_url \
      postgres_password keycloak_db_password kc_bootstrap_admin_password bff_client_secret \
      notify_client_secret members_client_secret
    ;;
  *) usage ;;
esac
