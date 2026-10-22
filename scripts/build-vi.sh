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

# BusyBox 1.00's vi.c explicitly supports a standalone build via -DSTANDALONE.
# This avoids linking the editor to BusyBox's multi-call runtime/libbb.
VI_VERSION=1.00
SOURCE_ARCHIVE="busybox-${VI_VERSION}.tar.bz2"
SOURCE_URL="https://busybox.net/downloads/$SOURCE_ARCHIVE"

mkdir -p "$ASSETS"
curl --fail --location --retry 3 "$SOURCE_URL" -o "$WORK/$SOURCE_ARCHIVE"
tar -xjf "$WORK/$SOURCE_ARCHIVE" -C "$WORK"

VI_SOURCE="$(find "$WORK/busybox-${VI_VERSION}" -type f -name vi.c -print -quit)"
test -n "$VI_SOURCE"
grep -q 'To compile for standalone use' "$VI_SOURCE"
grep -q '#define vi_main.*main' "$VI_SOURCE"

# Static musl linkage avoids a glibc-version dependency on the target machine.
musl-gcc -Wall -Os -s -static -DSTANDALONE -o "$WORK/vi" "$VI_SOURCE"
strip --strip-unneeded "$WORK/vi"
cp "$WORK/vi" "$ASSETS/vi"

test -x "$ASSETS/vi"
if readelf -l "$ASSETS/vi" | grep -q 'INTERP'; then
  echo "Expected a statically linked standalone vi: $ASSETS/vi" >&2
  exit 1
fi

echo "Standalone vi source: $VI_SOURCE"
sha256sum "$WORK/$SOURCE_ARCHIVE"
file "$ASSETS/vi"
stat -c 'Embedded vi size: %s bytes (%n)' "$ASSETS/vi"

if [[ -n "$SOURCE_OUT" ]]; then
  cp "$WORK/$SOURCE_ARCHIVE" "$SOURCE_OUT/"
  cp "$VI_SOURCE" "$SOURCE_OUT/vi.c"
  cp "$ROOT/scripts/build-vi.sh" "$SOURCE_OUT/"
  cp "$ROOT/THIRD_PARTY_NOTICES.md" "$SOURCE_OUT/"
  MUSL_COPYRIGHT=""
  for candidate in /usr/share/doc/musl/copyright /usr/share/doc/musl-dev/copyright; do
    if [[ -f "$candidate" ]]; then
      MUSL_COPYRIGHT="$candidate"
      break
    fi
  done
  if [[ -z "$MUSL_COPYRIGHT" ]]; then
    echo "Could not locate the musl copyright notice for the source bundle." >&2
    exit 1
  fi
  cp "$MUSL_COPYRIGHT" "$SOURCE_OUT/musl-copyright.txt"
  sha256sum "$WORK/$SOURCE_ARCHIVE" > "$SOURCE_OUT/${SOURCE_ARCHIVE}.sha256"
  cat > "$SOURCE_OUT/BUILD-INFO.txt" <<EOF
Embedded helper: standalone C vi executable
Source project: BusyBox
Source release: ${VI_VERSION} (2004)
Source file: vi.c, built with -DSTANDALONE
Source archive: ${SOURCE_ARCHIVE}
Linkage: static musl libc
The bundled editor is the older standalone implementation; it is not the
current BusyBox 1.37.0 vi applet and does not promise identical features.
EOF
fi
