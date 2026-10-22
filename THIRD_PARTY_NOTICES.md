# Third-party terminal editor

## BusyBox vi

The Linux amd64 release embeds the upstream BusyBox 1.37.0 multi-call executable with only the `vi` applet enabled. The Go application extracts this executable into a private temporary directory and runs `busybox vi FILE` as a child process, then removes the temporary directory after the editor exits.

BusyBox is licensed under GPL-2.0. The `lf-third-party-sources.tar.gz` release asset contains the exact upstream source archive, the generated BusyBox configuration, this notice, and the build script used to create the executable. The source archive includes the applicable license text.

Source archive: https://busybox.net/downloads/busybox-1.37.0.tar.bz2

## Rebuilding

On Linux amd64, install a C toolchain, make, curl, and bzip2, then run:

```bash
./scripts/build-internal-tools.sh
CGO_ENABLED=0 go build -trimpath -o lf ./
```

The pager is not bundled. The `i` command uses `set pager`, then `PAGER`, and falls back to a system `less` executable.
