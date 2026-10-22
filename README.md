# LF

[Documentation](doc.md)
| [Wiki](https://github.com/gokcehan/lf/wiki)
| [#lf:matrix.org](https://matrix.to/#/#lf:matrix.org) (with IRC bridge)

[![Go Build](https://github.com/gokcehan/lf/actions/workflows/go.yml/badge.svg)](https://github.com/gokcehan/lf/actions/workflows/go.yml)

`lf` ("list files") is a terminal file manager written in Go, inspired by
[ranger](https://github.com/ranger/ranger). It is designed to stay small,
scriptable, and fast.

![icons-and-border](https://github.com/user-attachments/assets/d5623462-05ab-4921-aeb3-d377d4732f9e)
![image-preview](https://github.com/user-attachments/assets/9ff42a21-19dd-42fa-8407-5d055aa9d561)
![new-features](https://github.com/user-attachments/assets/2a9a32aa-a764-438f-a810-e687e979dcee)

## Features

- Cross-platform: Linux, macOS, BSDs, and Windows
- Single native binary; Linux amd64 release embeds a statically linked upstream BusyBox vi executable
- Asynchronous filesystem and preview work to keep the UI responsive
- Server/client architecture for multiple connected instances
- Configurable commands, shell integration, keybindings, colors, and icons
- Vi-style navigation with Normal and Visual modes

## This fork

This branch keeps the upstream `lf` design while adding:

- Lazy filesystem metadata: ordinary directory entries can be displayed using `DirEntry` information without an immediate `Lstat` for every file.
- The `i` viewer uses an external pager: `set pager PROGRAM`, then `PAGER`, then `less` from `PATH`.
- `set editor internal` runs the genuine BusyBox `vi` applet; the Linux amd64 release bundles BusyBox with only the `vi` applet enabled.
- BusyBox `vi` takes over the terminal while running and returns to `lf` on exit; it is not a Go-language imitation.

The bundled BusyBox executable is built separately and embedded as data, not linked into the Go application. Its upstream source archive, exact build configuration, build script, and license information are provided in the release's `lf-third-party-sources.tar.gz` asset. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

## Installation

Pre-built Linux and Windows binaries are published in the
[releases](https://github.com/AmrUser-48/lf/releases).

To build from source, install [Go](https://go.dev/).

On Unix:

```bash
env CGO_ENABLED=0 go install -trimpath -ldflags="-s -w" github.com/AmrUser-48/lf@latest
```

On Windows `cmd`:

```cmd
set CGO_ENABLED=0
go install -trimpath -ldflags="-s -w" github.com/AmrUser-48/lf@latest
```

On Windows PowerShell:

```powershell
$env:CGO_ENABLED = '0'
go install -trimpath -ldflags="-s -w" github.com/AmrUser-48/lf@latest
```

## Usage

Run `lf` to start in the current directory.

The default internal shortcuts added by this fork are:

```
e       edit current file
i       view current file with the configured/system pager
```

Enable the bundled BusyBox `vi` editor with:

```
set editor internal
```

Linux amd64 release binaries contain a statically linked BusyBox executable with only the `vi` applet enabled. It is launched as a child process after `lf` suspends its screen, preserving BusyBox `vi`'s actual behavior. Other platforms use a system `busybox` executable from `PATH`. The pager is not bundled; it comes from `set pager`, `PAGER`, or the system `less` in `PATH`.

To rebuild the bundled programs from source on Linux amd64, install Go, a C compiler, make, curl, and bzip2, then run:

```bash
./scripts/build-internal-tools.sh
CGO_ENABLED=0 go build -trimpath -o lf ./
```

Set `PAGER` or `pager` explicitly when you want a different external pager.

Run `lf -help` for command-line options and `lf -doc` for the full
documentation.

See [etc](etc) for shell/editor integrations and example configuration files.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
