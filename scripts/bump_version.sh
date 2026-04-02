#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

version="$1"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "version must match MAJOR.MINOR.PATCH" >&2
  exit 1
fi

root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root_dir"

tag="v${version}"

printf '%s\n' "$version" > internal/buildinfo/VERSION.txt

VERSION="$version" TAG="$tag" ruby <<'RUBY'
path = "README.md"
text = File.read(path)
version = ENV.fetch("VERSION")
tag = ENV.fetch("TAG")
text.gsub!(/Current release: `v[0-9.]+`/, "Current release: `#{tag}`")
text.gsub!(/go install github\.com\/whosthatknocking\/kpx@v[0-9.]+/, "go install github.com/whosthatknocking/kpx@#{tag}")
text.gsub!(/kpx_[0-9.]+_darwin_arm64\.tar\.gz/, "kpx_#{version}_darwin_arm64.tar.gz")
text.gsub!(/kpx_[0-9.]+_darwin_arm64\/kpx/, "kpx_#{version}_darwin_arm64/kpx")
text.gsub!(/\.\/scripts\/bump_version\.sh [0-9.]+/, "./scripts/bump_version.sh #{version}")
File.write(path, text)
RUBY

./scripts/check_release.sh
echo "updated release version to ${tag}"
