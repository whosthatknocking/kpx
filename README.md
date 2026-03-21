# kpx

`kpx` is a focused Go CLI for working with KeePassXC-compatible `KDBX4` password databases.

It aims to be small, scriptable, and easy to audit:

- CLI only
- `KDBX4` only
- macOS-first
- safe atomic saves
- secrets redacted by default

## Status

`kpx` is usable today for password-only `KDBX4` workflows.

Current release: `v0.1.6`

## Overview

`kpx` is intentionally narrower than a full KeePass desktop replacement. The goal is a dependable command-line tool for people who want to:

- create and maintain local encrypted password databases
- script common vault operations
- inspect and update entries without leaving the terminal
- keep compatibility with KeePassXC workflows

Implemented today:

- create a new database
- validate and open an existing database
- list and create groups
- list, show, add, edit, and delete entries
- search entries by title
- store an optional default database in `~/.kpx/config.yml`
- support optional default `reveal` behavior in `~/.kpx/config.yml`
- support optional master password caching for a configured number of seconds
- back up the database before saving, with configurable destination and filename format
- omit the database argument for vault commands when a default is configured
- show the CLI version via `kpx version` and `kpx --version`
- atomic save behavior
- a refactored package layout with smaller command and vault files
- round-trip and fixture-based test coverage
- official on-demand KeePassXC compatibility fixtures
- advisory locking for cooperating `kpx` processes during reads and writes

## Install

Install the latest version with Go:

```bash
go install github.com/whosthatknocking/kpx@latest
```

Install a specific version:

```bash
go install github.com/whosthatknocking/kpx@v0.1.6
```

Build from source:

```bash
git clone https://github.com/whosthatknocking/kpx.git
cd kpx
go build -o kpx .
```

Requirements:

- Go `1.25` or newer for source builds
- macOS is the primary supported platform today
- Unix-style advisory locking is used for cooperating `kpx` processes

The base release version is stored in [`internal/buildinfo/VERSION.txt`](./internal/buildinfo/VERSION.txt). Builds always take the base version from that file, and append VCS metadata automatically when available.

## Quick Start

Create a vault:

```bash
printf '%s\n' 'master-password' | ./kpx db create ./vault.kdbx --password-stdin --name "Personal Vault"
```

Create a group:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin group add ./vault.kdbx /Personal
```

Add an entry:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin entry add ./vault.kdbx /Personal/GitHub \
  --username alice \
  --password 'entry-password' \
  --url https://github.com \
  --notes 'Personal account'
```

Show an entry:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin entry show ./vault.kdbx /Personal/GitHub
```

Reveal a password explicitly:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin entry show ./vault.kdbx /Personal/GitHub --reveal
```

Search by title:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin find ./vault.kdbx github
```

## Configuration

Create `~/.kpx/config.yml`:

```yaml
default_database: /Users/you/vault.kdbx
reveal: false
master_password_cache_seconds: 0
backup_directory: ""
backup_filename_format: "{db_stem}.{timestamp}.{db_ext}"
save_method: "temporary_file"
```

Set `master_password_cache_seconds` to a positive number to cache the master password for that many seconds. The default is `0`, which disables caching.

Leave `backup_directory` empty to store backups alongside the database. The default filename format uses the original database filename plus a UTC timestamp.

`save_method` defaults to `"temporary_file"`. Set it to `"direct_write"` to write directly to the database file instead.

Available backup filename placeholders:

- `{db_filename}`
- `{db_stem}`
- `{db_ext}`
- `{timestamp}`

Then omit the database path for vault commands:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin entry show /Personal/GitHub
```

If `reveal: true` is set in the config, `entry show` reveals passwords by default. Passing `--reveal` on the CLI still takes precedence when you want to override the config for a specific command.

## Commands

The CLI follows a simple noun/verb structure:

```bash
kpx db ...
kpx group ...
kpx entry ...
kpx find ...
```

Version checks:

```bash
kpx version
kpx --version
```

Available today:

- `kpx db create`
- `kpx db validate`
- `kpx group ls`
- `kpx group add`
- `kpx entry ls`
- `kpx entry show`
- `kpx entry add`
- `kpx entry edit`
- `kpx entry rm`
- `kpx find`
- `kpx version`
- `kpx --version`

Path rules:

- group paths look like `/Personal/Email`
- entry paths look like `/Personal/GitHub`
- the database argument is optional when `~/.kpx/config.yml` defines `default_database`
- `entry show` uses `reveal` from config unless `--reveal` is explicitly passed
- ambiguous matches fail closed

## Security Notes

- secrets are redacted by default
- password prompts use the controlling tty and do not echo
- master password caching is disabled by default
- when enabled, cached master passwords are stored on disk under `~/.kpx/master-password-cache.yml` with restrictive file permissions and a cooperating file lock
- database saves create a backup of the existing file before replacing it
- save method defaults to temporary-file-then-rename and can be changed to direct write in config
- cooperating `kpx` processes take advisory locks during database reads and hold an exclusive lock across write operations
- non-interactive usage supports stdin-based secrets
- writes are atomic
- destructive entry deletion requires confirmation unless `--force` is provided

Current limitations:

- key files are not supported yet
- JSON output is not supported yet
- advisory locks coordinate cooperating `kpx` processes, not every external KeePass tool
- `direct_write` is available for compatibility, but `temporary_file` remains the safer default

## Roadmap

Planned next:

- key file support
- password + key file support
- group rename, move, and delete
- JSON output
- stronger KeePassXC fixture coverage

## Compatibility

`kpx` is currently built around [`tobischo/gokeepasslib/v3`](https://pkg.go.dev/github.com/tobischo/gokeepasslib/v3) behind an internal adapter.

The test suite currently includes:

- CLI end-to-end flow tests
- repeated save/open round-trip tests
- atomic write tests
- fixture-driven compatibility tests across both save methods and backup-enabled saves

The compatibility harness includes official on-demand fixtures from the KeePassXC upstream repository, currently covering `NewDatabase.kdbx` and `NonAscii.kdbx` from the upstream test suite.

Remote fixtures are opt-in so regular `go test ./...` runs stay offline. To exercise them explicitly:

```bash
KPX_REMOTE_FIXTURES=1 go test ./internal/testcompat -run TestCompatibilityFixtures
```

For reliability, fixture URLs are pinned to a specific upstream commit and verified with SHA-256 instead of following a floating `latest` URL.

## Development

Development information starts here so the sections above stay focused on end users.

Run the test suite:

```bash
GOCACHE=/tmp/gocache go test ./...
```

Run additional static analysis:

```bash
GOCACHE=/tmp/gocache go vet ./...
```

Project layout:

```text
cmd/                Cobra command definitions
internal/buildinfo/ version metadata exposed to the CLI
internal/cli/       prompting, exit handling, and CLI helpers
internal/config/    optional user config for default database selection
internal/store/     atomic file writes
internal/testcompat/fixture-driven compatibility tests
internal/vault/     KDBX-backed vault adapter
```

Additional project docs:

## Non-Goals

Out of scope for the early versions:

- GUI
- browser integration
- desktop integration
- SSH agent integration
- auto-type
- sync and merge
- attachments and entry history
- `KDBX3` support

## License

MIT. See [LICENSE](./LICENSE).
