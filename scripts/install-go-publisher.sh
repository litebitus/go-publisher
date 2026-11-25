#!/usr/bin/env bash
# Installs go-publisher VERSION for RUNNER_OS_NAME and RUNNER_ARCHITECTURE, as
# GitHub Actions names them, into RUNNER_TEMP, checked against the release's
# checksums, and adds its directory to GITHUB_PATH.
set -euo pipefail

case "$RUNNER_OS_NAME" in
  Linux) goos=linux ;;
  macOS) goos=darwin ;;
  *) echo "go-publisher publishes no binary for runner OS $RUNNER_OS_NAME" >&2; exit 1 ;;
esac
case "$RUNNER_ARCHITECTURE" in
  X64) goarch=amd64 ;;
  ARM64) goarch=arm64 ;;
  *) echo "go-publisher publishes no binary for runner architecture $RUNNER_ARCHITECTURE" >&2; exit 1 ;;
esac
archive="go-publisher_${VERSION#v}_${goos}_${goarch}.tar.gz"
release_url="https://github.com/litebitus/go-publisher/releases/download/${VERSION}"
install_dir="$RUNNER_TEMP/go-publisher-${VERSION}"
mkdir -p "$install_dir"
curl --fail --silent --show-error --location --retry 3 \
  --output "$install_dir/$archive" "$release_url/$archive" \
  --output "$install_dir/checksums.txt" "$release_url/checksums.txt"
expected="$(awk -v archive="$archive" '$2 == archive { print $1 }' "$install_dir/checksums.txt")"
[ -n "$expected" ] || { echo "checksums.txt of go-publisher ${VERSION} names no $archive" >&2; exit 1; }
if command -v sha256sum >/dev/null; then
  actual="$(sha256sum "$install_dir/$archive" | awk '{ print $1 }')"
else
  actual="$(shasum -a 256 "$install_dir/$archive" | awk '{ print $1 }')"
fi
[ "$actual" = "$expected" ] ||
  { echo "$archive has SHA-256 $actual, and checksums.txt states $expected" >&2; exit 1; }
tar -xzf "$install_dir/$archive" -C "$install_dir" go-publisher
echo "$install_dir" >> "$GITHUB_PATH"
echo "installed go-publisher ${VERSION} for ${goos}/${goarch}"
