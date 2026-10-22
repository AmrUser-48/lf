#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
ASSETS="$ROOT/internaltools/assets"
SOURCE_OUT="${SOURCE_OUT:-}"
if [[ -n "$SOURCE_OUT" ]]; then
  if [[ "$SOURCE_OUT" != /* ]]; then
    SOURCE_OUT="$ROOT/$SOURCE_OUT"
  fi
  mkdir -p "$SOURCE_OUT"
  SOURCE_OUT="$(cd "$SOURCE_OUT" && pwd)"
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

BUSYBOX_VERSION=1.37.0
BUSYBOX_ARCHIVE="busybox-${BUSYBOX_VERSION}.tar.bz2"

mkdir -p "$ASSETS"
curl --fail --location --retry 3 "https://busybox.net/downloads/$BUSYBOX_ARCHIVE" -o "$WORK/$BUSYBOX_ARCHIVE"
tar -xjf "$WORK/$BUSYBOX_ARCHIVE" -C "$WORK"

cd "$WORK/busybox-${BUSYBOX_VERSION}"
# Configure only the vi applet; do not bundle unrelated BusyBox utilities.
make allnoconfig
if grep -q '^# CONFIG_VI is not set$' .config; then
  sed -i 's/^# CONFIG_VI is not set$/CONFIG_VI=y/' .config
elif ! grep -q '^CONFIG_VI=y$' .config; then
  printf '\nCONFIG_VI=y\n' >> .config
fi
if grep -q '^# CONFIG_STATIC is not set$' .config; then
  sed -i 's/^# CONFIG_STATIC is not set$/CONFIG_STATIC=y/' .config
elif grep -q '^CONFIG_STATIC=' .config; then
  sed -i 's/^CONFIG_STATIC=.*/CONFIG_STATIC=y/' .config
else
  printf '\nCONFIG_STATIC=y\n' >> .config
fi
# BusyBox has additional defaulted Kconfig symbols (e.g. FEATURE_VI_MAX_LEN).
# Feed empty answers so oldconfig selects each symbol's documented default.
make oldconfig < <(yes '')
grep -q '^CONFIG_VI=y$' .config
grep -q '^CONFIG_STATIC=y$' .config
make -j"$(nproc)"
./busybox --list | grep -qx vi
test "$(./busybox --list | wc -l)" -eq 1
strip --strip-unneeded ./busybox
cp ./busybox "$ASSETS/busybox"

test -x "$ASSETS/busybox"
if readelf -l "$ASSETS/busybox" | grep -q 'INTERP'; then
  echo "Expected a statically linked BusyBox: $ASSETS/busybox" >&2
  exit 1
fi
file "$ASSETS/busybox"
ls -lh "$ASSETS/busybox"

if [[ -n "$SOURCE_OUT" ]]; then
  cp "$WORK/$BUSYBOX_ARCHIVE" "$SOURCE_OUT/"
  cp "$WORK/busybox-${BUSYBOX_VERSION}/.config" "$SOURCE_OUT/busybox-${BUSYBOX_VERSION}.config"
  cp "$ROOT/scripts/build-internal-tools.sh" "$SOURCE_OUT/"
  cp "$ROOT/THIRD_PARTY_NOTICES.md" "$SOURCE_OUT/"
  cat > "$SOURCE_OUT/BUILD-INFO.txt" <<EOF
Embedded helper built on Linux amd64.
BusyBox version: ${BUSYBOX_VERSION}
BusyBox source archive: $BUSYBOX_ARCHIVE
BusyBox applet configuration: vi only
BusyBox configuration: busybox-${BUSYBOX_VERSION}.config
EOF
fi
