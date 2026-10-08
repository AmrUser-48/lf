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
- Single native binary with no runtime dependencies
- Asynchronous filesystem and preview work to keep the UI responsive
- Server/client architecture for multiple connected instances
- Configurable commands, shell integration, keybindings, colors, and icons
- Vi-style navigation with Normal and Visual modes

## This fork

This branch keeps the upstream `lf` design while adding:

- Lazy filesystem metadata: ordinary directory entries can be displayed using `DirEntry` information without an immediate `Lstat` for every file.
- Delayed file metadata in the bottom status line, fetched after the cursor remains on a file for 1.5 seconds.
- Built-in less-style text viewer on `i`. When no explicit pager is configured it stays inside `lf`; `set pager less` or `PAGER=less` uses external `less`.
- Pager navigation with `j/k`, arrows, PageUp/PageDown, Space/`b`, `g/G`, and Alt+`j/k/f/b`.
- Optional internal vi editor on `e` with `set editor internal`. It follows BusyBox `vi` command conventions, including `ZZ` save-and-exit and `ZQ` discard-and-exit. `F10` is also a discard-and-exit shortcut.
- The normal external editor remains available through `$EDITOR`, including BusyBox `vi`.

The internal editor is a fresh Go implementation of vi-style behavior; BusyBox source is not included in this MIT-licensed fork.

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
i       view current file
ZZ      save and exit internal vi
ZQ      discard and exit internal vi
F10     discard and exit internal vi

Alt+j   scroll viewer down one line
Alt+k   scroll viewer up one line
Alt+F   viewer page down
Alt+B   viewer page up
```

Enable the internal editor with:

```
set editor internal
```

Leave the editor unset (or set it to `external`) to keep using `$EDITOR`.
Set `PAGER` or `pager` explicitly when you want an external pager.

Run `lf -help` for command-line options and `lf -doc` for the full
documentation.

See [etc](etc) for shell/editor integrations and example configuration files.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
