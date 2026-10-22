# Third-party terminal tools

The Linux amd64 release embeds two upstream executables as data and launches them as separate child processes. Their C code is not linked into the Go program.

## BusyBox vi

The editor payload is built from BusyBox 1.37.0. The BusyBox project is licensed under GPL-2.0. The build script uses the upstream release archive at https://busybox.net/downloads/busybox-1.37.0.tar.bz2 and records the exact BusyBox build configuration in the release source bundle. The bundled executable is the BusyBox multi-call binary; lf invokes its `vi` applet.

## GNU less

The pager payload is built from GNU less 710, released on 17 September 2026. Less is distributed under either the GNU GPL version 3 or later, or the separate Less License; see the `LICENSE` and `COPYING` files in the upstream source archive. Source archive: https://www.greenwoodsoftware.com/less/less-710.tar.gz

## Rebuilding

The release's `lf-third-party-sources.tar.gz` asset contains the exact upstream source archives, the BusyBox `.config`, this notice, and the helper build script. Run `scripts/build-internal-tools.sh` on Linux amd64 with a C toolchain, make, curl, bzip2, and ncurses development files to rebuild the embedded payloads. The Go program extracts them into a private temporary directory for execution and removes that directory when the child exits.
