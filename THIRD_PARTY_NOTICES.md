# Third-party terminal editor

## BusyBox standalone vi.c

The Linux amd64 release embeds a separate, statically linked executable compiled from the standalone `vi.c` in BusyBox 1.00 (released in 2004). The source explicitly supports standalone compilation through `-DSTANDALONE`; it does not require BusyBox's multi-call runtime or `libbb`. This old implementation is not feature-identical to the current BusyBox `vi` applet.

The editor source was written by Sterling Huxley and is licensed under GPL-2.0-or-later, as stated in the source. The `lf-third-party-sources.tar.gz` release asset includes the complete upstream BusyBox 1.00 source archive, the extracted `vi.c`, the build script, and this notice.

Source archive: https://busybox.net/downloads/busybox-1.00.tar.bz2

## Static C runtime

The editor is statically linked against musl libc to avoid a target glibc version dependency. musl is MIT-licensed; the source bundle includes the musl copyright notice provided by the build environment.

## Rebuilding

On Linux amd64, install a C toolchain, make, curl, bzip2, and musl-tools, then run:

```bash
./scripts/build-vi.sh
CGO_ENABLED=0 go build -trimpath -o lf ./
```

The pager is not bundled. The `i` command uses `set pager`, then `PAGER`, and falls back to a system `less` executable.
