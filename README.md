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
- Single native binary; Linux amd64 release embeds a statically linked standalone C `vi` executable
- Asynchronous filesystem and preview work to keep the UI responsive
- Server/client architecture for multiple connected instances
- Configurable commands, shell integration, keybindings, colors, and icons
- Vi-style navigation with Normal and Visual modes

## This fork

This branch keeps the upstream `lf` design while adding:

- Lazy filesystem metadata: ordinary directory entries can be displayed using `DirEntry` information without an immediate `Lstat` for every file.
- The `i` viewer uses an external pager: `set pager PROGRAM`, then `PAGER`, then `less` from `PATH`.
- `set editor internal` launches the standalone upstream BusyBox `vi.c` implementation, compiled as its own executable rather than as a BusyBox multi-call binary.
- The bundled editor is C, not a Go imitation. It is the older BusyBox 1.00 (2004) standalone implementation, so its feature set differs from the current BusyBox applet.

The Linux amd64 `lf` release embeds the static `vi` executable and extracts it to a private temporary directory only while the editor runs. The release's `lf-third-party-sources.tar.gz` asset contains the upstream source archive, standalone source file, build script, and license notices. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

## Installation

Pre-built Linux and Windows binaries are published in the
[releases](https://github.com/AmrUser-48/lf/releases).

To rebuild the Linux amd64 binary with its embedded standalone editor, install Go, a C compiler, make, curl, bzip2, and musl-tools, then run:

```bash
./scripts/build-vi.sh
CGO_ENABLED=0 go build -trimpath -o lf ./
```

On Unix, `CGO_ENABLED=0 go install -trimpath -ldflags="-s -w" github.com/AmrUser-48/lf@latest` builds `lf` without embedding the generated editor unless you run the helper build script first. In that case, the internal editor falls back to a system `vi`, or BusyBox `vi` if available.

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

Enable the internal editor with:

```
set editor internal
```

Linux amd64 release binaries contain the statically linked standalone `vi` executable. Other platforms use `vi` from `PATH`, falling back to `busybox vi` if available. The pager is not bundled; it comes from `set pager`, `PAGER`, or the system `less` in `PATH`.

Run `lf -help` for command-line options and `lf -doc` for the full
documentation.

See [etc](etc) for shell/editor integrations and example configuration files.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
