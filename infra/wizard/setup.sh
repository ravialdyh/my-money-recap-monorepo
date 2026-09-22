#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../.."

SECRETS_DIR=${WIZARD_SECRETS_DIR:-infra/secrets}
EXAMPLE=$SECRETS_DIR/example.yaml
SECRETS=$SECRETS_DIR/prod.yaml
RULES=${WIZARD_RULES_FILE:-.sops.yaml}
KEYSET=infra/secrets/keyset.sh
KEY_FILE=${SOPS_AGE_KEY_FILE:-$HOME/.config/sops/age/keys.txt}
STACK_KEYS=$("$KEYSET" stack)
STACK_TITLE="Values the stack itself needs, chosen once and assembled"
PLAIN=

STEPS=(
"@decrypt|One age key pair, the only secret a human holds|age-keygen here, free and local"
"@byhand|Apple Developer Program enrolment|developer.apple.com, USD 99 a year"
"app_store_connect_api_key|App Store Connect API key|App Store Connect > Users and Access > Integrations"
"apple_services_id apple_team_id apple_sign_in_key_id apple_sign_in_key apple_push_key|Apple Services ID and the two .p8 keys|Apple Developer > Identifiers (the Services ID), Membership details (the Team ID), then Keys (the Sign in with Apple .p8 and its Key ID, then the APNs .p8)"
"google_client_id google_client_secret|Google OAuth client|Google Auth Platform > Clients"
"open_exchange_rates_app_id|Open Exchange Rates App ID|the Open Exchange Rates account dashboard"
"smtp_token|SMTP provider account and its first token|the provider's dashboard"
"domain registrar_token|Domain, and a registrar token scoped to its zone|the registrar's dashboard"
"cloud_api_token deploy_ssh_private_key|Cloud account, its API token and a deploy key|the cloud dashboard, then ssh-keygen here"
)

fail() { echo "FAIL: $*" >&2; exit 2; }

usage() {
  printf 'usage: %s [--check]\n\n' "$0"
  printf '  no argument  walk the nine steps in order\n'
  printf '  --check      report which are done, non-zero while any checkable one is not\n'
}

example_keys() { "$KEYSET" keys "$EXAMPLE"; }

placeholder() { "$KEYSET" value "$EXAMPLE" "$1"; }

step_keys() {
  local record
  for record in "${STEPS[@]}"; do
    case ${record%%|*} in
      @*) ;;
      *) tr ' ' '\n' <<<"${record%%|*}" ;;
    esac
  done
}

claimed_keys() { step_keys; tr ' ' '\n' <<<"$STACK_KEYS"; }

missing_from() {
  local candidate haystack
  haystack=$'\n'$2$'\n'
  for candidate in $1; do
    case $haystack in
      *$'\n'"$candidate"$'\n'*) ;;
      *) echo "$candidate" ;;
    esac
  done
}

join() {
  local out='' item
  for item in $1; do out="${out:+$out, }$item"; done
  printf '%s' "$out"
}

key_state() {
  if ! jq -e --arg k "$1" 'has($k)' >/dev/null <<<"$PLAIN"; then
    echo absent
  elif [ "$(jq -r --arg k "$1" '.[$k]' <<<"$PLAIN")" = "$(placeholder "$1")" ]; then
    echo unfilled
  else
    echo filled
  fi
}

report() {
  if [ -n "$4" ]; then
    printf '%2s  %-7s  %s — %s\n' "$1" "$2" "$3" "$4"
  else
    printf '%2s  %-7s  %s\n' "$1" "$2" "$3"
  fi
}

value_note() {
  local key unfilled='' absent='' note=''
  for key in $1; do
    case $(key_state "$key") in
      absent) absent="$absent $key" ;;
      unfilled) unfilled="$unfilled $key" ;;
    esac
  done
  if [ -n "$unfilled" ]; then note=$(join "$unfilled"); fi
  if [ -n "$absent" ]; then note="${note:+$note; }$(join "$absent") absent from $SECRETS"; fi
  printf '%s' "$note"
}

drift() {
  local key problems=0
  for key in $(missing_from "$(example_keys)" "$(claimed_keys)"); do
    echo "drift: $EXAMPLE declares $key, which no step and no stack value produces"
    problems=$((problems + 1))
  done
  for key in $(missing_from "$(step_keys)" "$(example_keys)"); do
    echo "drift: a step produces $key, which $EXAMPLE does not declare"
    problems=$((problems + 1))
  done
  return "$problems"
}

check() {
  [ -f "$EXAMPLE" ] || fail "no key set at $EXAMPLE"
  local blocked='' record keys title where index=0 finished=0 pending=0 byhand=0 note

  if [ ! -f "$SECRETS" ]; then
    blocked="$SECRETS does not exist yet, so no value in it can be read"
  elif ! PLAIN=$(sops decrypt --output-type json "$SECRETS" 2>/dev/null); then
    blocked="$SECRETS does not decrypt with the age key this machine holds"
  fi
  if [ -n "$blocked" ]; then printf '%s.\n\n' "$blocked"; fi

  for record in "${STEPS[@]}"; do
    index=$((index + 1))
    IFS='|' read -r keys title where <<<"$record"
    case $keys in
      @byhand)
        report "$index" "by hand" "$title" "$where, and never machine-checked"
        byhand=$((byhand + 1))
        ;;
      @decrypt)
        if [ -n "$blocked" ]; then
          report "$index" missing "$title" "no file this machine can decrypt"
          pending=$((pending + 1))
        else
          report "$index" "done" "$title" "$SECRETS decrypts here"
          finished=$((finished + 1))
        fi
        ;;
      *)
        if [ -n "$blocked" ]; then note=$(join "$keys"); else note=$(value_note "$keys"); fi
        if [ -n "$note" ]; then
          report "$index" missing "$title" "$note"
          pending=$((pending + 1))
        else
          report "$index" "done" "$title" ""
          finished=$((finished + 1))
        fi
        ;;
    esac
  done

  local stack problems=0
  if [ -n "$blocked" ]; then stack=$(join "$STACK_KEYS"); else stack=$(value_note "$STACK_KEYS"); fi
  if [ -n "$stack" ]; then
    report - missing "$STACK_TITLE" "$stack"
  else
    report - "done" "$STACK_TITLE" ""
  fi

  drift || problems=$?
  printf '\n%d of %d checkable steps done, %d reported by hand, stack values %sfilled, %d drift.\n' \
    "$finished" "$((finished + pending))" "$byhand" "${stack:+not }" "$problems"
  [ "$pending" -eq 0 ] && [ -z "$stack" ] && [ "$problems" -eq 0 ]
}

ask() {
  local reply
  printf '  %s? [y/N] ' "$1"
  read -r reply || return 1
  case $reply in [yY]*) return 0 ;; *) return 1 ;; esac
}

age_step() {
  local recipient
  umask 077
  if [ -f "$KEY_FILE" ]; then
    printf '  %s already holds a key, and the wizard never overwrites one.\n' "$KEY_FILE"
  elif ask "Generate the age key pair now"; then
    mkdir -p "$(dirname "$KEY_FILE")"
    age-keygen -o "$KEY_FILE"
  else
    return 0
  fi
  recipient=$(age-keygen -y "$KEY_FILE")
  grep -qsF "$recipient" "$RULES" ||
    printf 'creation_rules:\n  - path_regex: infra/secrets/prod\\.yaml$\n    age: %s\n' "$recipient" >"$RULES"
  if [ -f "$SECRETS" ]; then
    printf '  %s already exists, and the wizard never overwrites it either.\n' "$SECRETS"
  else
    sops encrypt --age "$recipient" --input-type yaml --output-type yaml \
      --filename-override "$SECRETS" --output "$SECRETS" "$EXAMPLE"
    printf '  %s written from %s.\n' "$SECRETS" "$EXAMPLE"
  fi
  printf '  %s names the key %s holds.\n' "$RULES" "$KEY_FILE"
  printf '  Copy %s to the GitHub production environment; the machine never needs it.\n' "$KEY_FILE"
}

ssh_step() {
  local out=${TMPDIR:-/tmp}/mmr-deploy-key
  if [ -f "$out" ]; then
    printf '  %s already exists; paste it, then delete it.\n' "$out"
    return 0
  fi
  ask "Generate the deploy SSH key pair now" || return 0
  ssh-keygen -q -t ed25519 -N '' -C mmr-deploy -f "$out"
  printf '  Public half, which cloud-init takes:\n  %s\n' "$(cat "$out.pub")"
  printf '  Paste %s into deploy_ssh_private_key, then delete it.\n' "$out"
}

value_step() {
  printf '  Fills: %s\n' "$(join "$1")"
  if [ ! -f "$SECRETS" ]; then
    printf '  %s does not exist yet; do step 1 first.\n' "$SECRETS"
    return 0
  fi
  case " $1 " in *" deploy_ssh_private_key "*) ssh_step ;; esac
  ask "Open sops edit now" || return 0
  sops edit "$SECRETS"
}

walk() {
  local record keys title where index=0
  for record in "${STEPS[@]}"; do
    index=$((index + 1))
    IFS='|' read -r keys title where <<<"$record"
    printf '\nStep %d of %d: %s\n  Where: %s\n' "$index" "${#STEPS[@]}" "$title" "$where"
    case $keys in
      @decrypt) age_step ;;
      @byhand) printf '  Nothing to paste. Check mode reports this one by hand.\n' ;;
      *) value_step "$keys" ;;
    esac
  done
  printf '\nRun %s --check to see what is still missing.\n' "$0"
}

main() {
  case ${1:-} in
    --check) check ;;
    -h | --help) usage ;;
    "") walk ;;
    *) usage >&2; exit 2 ;;
  esac
}

main "$@"
