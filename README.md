# wips-cli

> **Local-first Event Logger for Developers**

`wips-cli` is a lightweight CLI tool designed to capture your development context—notes, git commits, and more—stored securely on your local machine. It serves as a personal "flight recorder" for your coding sessions.

### What is "wips"?

**wips** is simply the plural form of **WIP** (Work In Progress).
It captures the stream of small intermediate states and thoughts during development that are rarely preserved in commit history.

## Features

- 🔒 **Local-First**: All data is stored locally in NDJSON/JSON format. No cloud dependencies.
- 🧠 **Context-Aware**: Automatically captures environment info, git repository details, and current working directory with every note.
- 🎣 **Git Integration**: Seamlessly hooks into `git commit` to log every commit with diff stats automatically.
- ⚡ **Zero Latency**: Written in Go for instant startup and minimal overhead.

## Installation

### Option 1: Go Install (Recommended)

Requires Go 1.25+ installed.

```bash
go install github.com/rynskrmt/wips-cli/cmd/wip@latest
```

Make sure your `GOPATH/bin` is in your `PATH`:

```bash
# zsh example
export PATH=$PATH:$(go env GOPATH)/bin
```

### Option 2: Manual Build

```bash
git clone https://github.com/rynskrmt/wips-cli.git
cd wips-cli
go build -o bin/wip ./cmd/wip

# Move to a directory in your PATH
sudo mv bin/wip /usr/local/bin/
```

### Option 3: Using Makefile

If you have `make` installed:

```bash
make dev # Build and install to $GOPATH/bin (No sudo required)
```

## Usage

### Basic Commands

| Command           | Alias | Description                              | Example                            |
| :---------------- | :---- | :--------------------------------------- | :--------------------------------- |
| `wip <msg>`       | -     | Record a note with auto-detected context | `wip "Refactoring auth logic"`     |
| `wip search <q>`  | -     | Search for events matching query         | `wip search "auth bug"`            |
| `wip edit [id]`   | `e`   | Edit an event (default: latest)          | `wip e` / `wip e 01J...`           |
| `wip delete [id]` | -     | Delete an event (default: latest)        | `wip delete` / `wip delete 01J...` |
| `wip undo`        | `u`   | Undo the last recorded event             | `wip u`                            |
| `wip summary`     | `sum` | Show summary of events (today/week)      | `wip sum --range=week`             |

### Viewing Logs (`tail`)

`wip tail` (alias: `wip t`) displays recent events. By default, it **only shows logs for the current directory** (and subdirectories).

| Command             | Alias  | Description                               |
| :------------------ | :----- | :---------------------------------------- |
| `wip tail`          | `t`    | Show events for current directory context |
| `wip tail --global` | `t -g` | Show ALL events (ignore context)          |
| `wip tail --id`     | -      | Show events with IDs                      |
| `wip tail -n 20`    | -      | Show last 20 events                       |

### Git Integration

To enable automatic commit logging, run this inside your repository:

```bash
wip hooks install
```

Once installed, every `git commit` will be automatically logged to `wip`.




## Data Storage

Your data is stored in:
- **macOS**: `~/Library/Application Support/wip/`
- **Linux**: `~/.local/share/wip/` (XDG standard)
- **Windows**: `%APPDATA%\wip\`

## License

MIT License

