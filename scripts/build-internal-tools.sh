#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
ASSETS="$ROOT/internaltools/assets"
if [[ -n "${SOURCE_OUT:-}" ]]; then
  if [[ "$SOURCE_OUT" != /* ]]; then
    SOURCE_OUT="$ROOT/$SOURCE_OUT"
  fi
  mkdir -p "$SOURCE_OUT"
  SOURCE_OUT="$(cd "$SOURCE_OUT" && pwd)"
fi
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

BUSYBOX_VERSION=1.37.0
LESS_VERSION=710
BUSYBOX_ARCHIVE="busybox-${BUSYBOX_VERSION}.tar.bz2"
LESS_ARCHIVE="less-${LESS_VERSION}.tar.gz"

mkdir -p "$ASSETS"
curl --fail --location --retry 3 "https://busybox.net/downloads/$BUSYBOX_ARCHIVE" -o "$WORK/$BUSYBOX_ARCHIVE"
curl --fail --location --retry 3 "https://www.greenwoodsoftware.com/less/$LESS_ARCHIVE" -o "$WORK/$LESS_ARCHIVE"

tar -xjf "$WORK/$BUSYBOX_ARCHIVE" -C "$WORK"
tar -xzf "$WORK/$LESS_ARCHIVE" -C "$WORK"

cd "$WORK/busybox-${BUSYBOX_VERSION}"
make defconfig
if grep -q '^# CONFIG_STATIC is not set$' .config; then
  sed -i 's/^# CONFIG_STATIC is not set$/CONFIG_STATIC=y/' .config
elif grep -q '^CONFIG_STATIC=' .config; then
  sed -i 's/^CONFIG_STATIC=.*/CONFIG_STATIC=y/' .config
else
  printf '\nCONFIG_STATIC=y\n' >> .config
fi
make oldconfig </dev/null
grep -q '^CONFIG_VI=y$' .config
grep -q '^CONFIG_STATIC=y$' .config
make -j"$(nproc)"
./busybox --list | grep -qx vi
strip --strip-unneeded ./busybox
cp ./busybox "$ASSETS/busybox"

cd "$WORK/less-${LESS_VERSION}"
LDFLAGS="-static" ./configure --with-regex=posix
make -j"$(nproc)"
strip --strip-unneeded ./less
cp ./less "$ASSETS/less"

for bin in "$ASSETS/busybox" "$ASSETS/less"; do
  test -x "$bin"
  if readelf -l "$bin" | grep -q 'INTERP'; then
    echo "Expected a statically linked helper: $bin" >&2
    exit 1
  fi
  file "$bin"
  ls -lh "$bin"
done

if [[ -n "${SOURCE_OUT:-}" ]]; then
  mkdir -p "$SOURCE_OUT"
  SOURCE_OUT="$(cd "$SOURCE_OUT" && pwd)"
  cp "$WORK/$BUSYBOX_ARCHIVE" "$SOURCE_OUT/"
  cp "$WORK/$LESS_ARCHIVE" "$SOURCE_OUT/"
  cp "$WORK/busybox-${BUSYBOX_VERSION}/.config" "$SOURCE_OUT/busybox-${BUSYBOX_VERSION}.config"
  cp "$ROOT/scripts/build-internal-tools.sh" "$SOURCE_OUT/"
  cp "$ROOT/THIRD_PARTY_NOTICES.md" "$SOURCE_OUT/"
  cat > "$SOURCE_OUT/BUILD-INFO.txt" <<EOF
Embedded helpers built on Linux amd64.
BusyBox version: ${BUSYBOX_VERSION}
GNU less version: ${LESS_VERSION}
BusyBox source archive: $BUSYBOX_ARCHIVE
GNU less source archive: $LESS_ARCHIVE
BusyBox configuration: busybox-${BUSYBOX_VERSION}.config
EOF
fi
