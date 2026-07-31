#!/usr/bin/env sh

set -eu

die() {
  printf 'spotui-install: %s\n' "$1" >&2
  exit 1
}

command -v curl >/dev/null 2>&1 || die "curl is required"

if command -v sha256sum >/dev/null 2>&1; then
  checksum_tool=sha256sum
elif command -v shasum >/dev/null 2>&1; then
  checksum_tool=shasum
else
  die "sha256sum or shasum is required"
fi

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) die "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) die "unsupported architecture: $(uname -m)" ;;
esac

if [ "$os" = linux ] && [ "$arch" = arm64 ]; then
  die "Linux arm64 is not published yet"
fi

home_dir=${HOME:-}
[ -n "$home_dir" ] || die "HOME is not set; use SPOTUI_INSTALL_DIR explicitly"

repo=${SPOTUI_REPO:-davicbtoliveira/spotui}
install_dir=${SPOTUI_INSTALL_DIR:-$home_dir/.local/bin}
version=${SPOTUI_VERSION:-latest}

if [ "$os" = linux ]; then
  format=appimage
  asset=spotui-linux-amd64.AppImage
else
  format=archive
  asset="spotui-${os}-${arch}.tar.gz"
  command -v tar >/dev/null 2>&1 || die "tar is required on macOS"
fi

if [ "$version" = latest ]; then
  release_base="https://github.com/$repo/releases/latest/download"
else
  case "$version" in
    v*) tag=$version ;;
    *) tag="v$version" ;;
  esac
  release_base="https://github.com/$repo/releases/download/$tag"
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

curl --fail --location --silent --show-error \
  "$release_base/$asset" --output "$tmp_dir/$asset"
curl --fail --location --silent --show-error \
  "$release_base/SHA256SUMS" --output "$tmp_dir/SHA256SUMS"

expected_checksum=$(awk -v asset="$asset" \
  '$2 == asset || $2 == "*" asset { print $1; exit }' \
  "$tmp_dir/SHA256SUMS")
[ -n "$expected_checksum" ] || die "checksum for $asset was not found"

if [ "$checksum_tool" = sha256sum ]; then
  actual_checksum=$(sha256sum "$tmp_dir/$asset" | awk '{print $1}')
else
  actual_checksum=$(shasum -a 256 "$tmp_dir/$asset" | awk '{print $1}')
fi

[ "$expected_checksum" = "$actual_checksum" ] || \
  die "checksum verification failed for $asset"

if [ "$format" = appimage ]; then
  binary="$tmp_dir/$asset"
else
  mkdir "$tmp_dir/extracted"
  tar -xzf "$tmp_dir/$asset" -C "$tmp_dir/extracted"
  binary="$tmp_dir/extracted/spotui-${os}-${arch}/spotui"
fi
[ -f "$binary" ] || die "the release artifact does not contain spotui"

mkdir -p "$install_dir"
if command -v install >/dev/null 2>&1; then
  install -m 755 "$binary" "$install_dir/spotui"
else
  cp "$binary" "$install_dir/spotui"
  chmod 755 "$install_dir/spotui"
fi

printf 'SpotUI installed at %s\n' "$install_dir/spotui"
case ":${PATH:-}:" in
  *:"$install_dir":*) ;;
  *) printf 'Add %s to PATH to run spotui from any terminal.\n' "$install_dir" ;;
esac
