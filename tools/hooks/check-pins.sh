#!/bin/sh
set -eu
cd "$(git rev-parse --show-toplevel)"

ci=.github/workflows/ci.yml
deploy=.github/workflows/deploy.yml
fail=0
disagree() {
  printf '%s\n' "$@" >&2
  fail=1
}

hook=$(awk '/repo: https:\/\/github\.com\/gitleaks\/gitleaks$/ { getline; print $2; exit }' .pre-commit-config.yaml)
release=$(sed -n 's|.*/gitleaks/releases/download/\([^/]*\)/.*|\1|p' .github/workflows/security.yml)
tarball=$(sed -n 's|.*/gitleaks_\([^_]*\)_linux_x64\.tar\.gz.*|\1|p' .github/workflows/security.yml)
if [ -z "$hook" ] || [ "$release" != "$hook" ] || [ "v$tarball" != "$hook" ]; then
  disagree "gitleaks pins disagree: .pre-commit-config.yaml has ${hook:-nothing}, security.yml downloads release ${release:-nothing}, tarball ${tarball:-nothing}" \
    "update the URL and the sha256 in .github/workflows/security.yml to the hook's version"
fi

sops_url() { sed -n 's|.*\(https://github\.com/getsops/sops/releases/download/[^ ]*\).*|\1|p' "$1"; }
sops_sha() { sed -n 's|.*"\([0-9a-f]\{64\}\)  [^"]*sops".*|\1|p' "$1"; }
if [ -z "$(sops_url "$ci")" ] || [ "$(sops_url "$ci")" != "$(sops_url "$deploy")" ] ||
  [ -z "$(sops_sha "$ci")" ] || [ "$(sops_sha "$ci")" != "$(sops_sha "$deploy")" ]; then
  disagree "sops pins disagree: $ci and $deploy must download the same URL with the same sha256"
fi

image=$(sed -n 's|.*ghcr\.io/opentofu/opentofu:\([^ ]*\).*|\1|p' infra/compose/compose.yaml)
zip=$(sed -n 's|.*/opentofu/releases/download/v\([^/]*\)/.*|\1|p' "$ci")
if [ -z "$image" ] || [ "$image" != "$zip" ]; then
  disagree "OpenTofu pins disagree: compose.yaml's realm-apply runs ${image:-nothing}, $ci installs ${zip:-nothing}"
fi

exit "$fail"
