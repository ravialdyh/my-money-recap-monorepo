#!/bin/sh
set -eu
cd "$(git rev-parse --show-toplevel)"

hook=$(awk '/repo: https:\/\/github\.com\/gitleaks\/gitleaks$/ { getline; print $2; exit }' .pre-commit-config.yaml)
release=$(sed -n 's|.*/gitleaks/releases/download/\([^/]*\)/.*|\1|p' .github/workflows/security.yml)
tarball=$(sed -n 's|.*/gitleaks_\([^_]*\)_linux_x64\.tar\.gz.*|\1|p' .github/workflows/security.yml)

if [ -n "$hook" ] && [ "$release" = "$hook" ] && [ "v$tarball" = "$hook" ]; then
  exit 0
fi
printf 'gitleaks pins disagree: .pre-commit-config.yaml has %s, security.yml downloads release %s, tarball %s\n' \
  "${hook:-nothing}" "${release:-nothing}" "${tarball:-nothing}" >&2
printf 'update the URL and the sha256 in .github/workflows/security.yml to the hook'\''s version\n' >&2
exit 1
