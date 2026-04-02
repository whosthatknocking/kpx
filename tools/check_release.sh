#!/usr/bin/env bash
set -euo pipefail

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

version="$(tr -d '[:space:]' < internal/buildinfo/VERSION.txt)"
tag="v${version}"

check_contains() {
  local file="$1"
  local needle="$2"
  if ! grep -Fq "$needle" "$file"; then
    echo "missing expected text in ${file}: ${needle}" >&2
    exit 1
  fi
}

check_contains README.md "Current release: \`${tag}\`"
check_contains README.md "go install github.com/whosthatknocking/kpx@${tag}"
check_contains README.md "tar -xzf kpx_${version}_darwin_arm64.tar.gz"
check_contains README.md "install -m 0755 kpx_${version}_darwin_arm64/kpx ~/.local/bin/kpx"
check_contains main_test.go "\"Tool Version: ${version}\","
check_contains main_test.go "if got := buildinfo.BaseVersion(); got != \"${version}\" {"

if [[ "${1:-}" == "--expect-tag" ]]; then
  current_tag="$(git describe --tags --exact-match 2>/dev/null || true)"
  if [[ "${current_tag}" != "${tag}" ]]; then
    echo "current commit tag ${current_tag:-<none>} does not match ${tag}" >&2
    exit 1
  fi
fi

echo "release metadata ok for ${tag}"
