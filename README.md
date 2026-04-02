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

Current release: `v0.1.9`

The project is still maturing, and the CLI surface, output details, config behavior, and internal Go APIs may change between early releases.

## Overview

`kpx` is intentionally narrower than a full KeePass desktop replacement. The goal is a dependable command-line tool for people who want to:

- create and maintain local encrypted password databases
- script common vault operations
- inspect and update entries without leaving the terminal
- keep compatibility with KeePassXC workflows

Desktop GUI work is out of scope for this project. `kpx` is intended to remain a CLI-first tool rather than grow into a desktop application.

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
- export a printable plaintext recovery document with secrets for secure paper backup
- emit JSON output for supported commands with `--json`
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
go install github.com/whosthatknocking/kpx@v0.1.9
```

Build from source:

```bash
git clone https://github.com/whosthatknocking/kpx.git
cd kpx
go build -o kpx .
```

Install from a release archive:

```bash
tar -xzf kpx_0.1.9_darwin_arm64.tar.gz
install -m 0755 kpx_0.1.9_darwin_arm64/kpx ~/.local/bin/kpx
```

Choose the archive that matches your platform:

- `kpx_<version>_darwin_amd64.tar.gz`
- `kpx_<version>_darwin_arm64.tar.gz`
- `kpx_<version>_linux_amd64.tar.gz`

Requirements:

- Go `1.25` or newer for source builds
- macOS is the primary supported platform today
- Linux release archives are published for `amd64`
- Unix-style advisory locking is used for cooperating `kpx` processes
- Non-Unix platforms may compile for development purposes, but real vault operations are not supported there today because advisory file locking is Unix-only in the current implementation

The base release version is stored in [`internal/buildinfo/VERSION.txt`](./internal/buildinfo/VERSION.txt). Builds always take the base version from that file, and append VCS metadata automatically when available.

## Shell Completion

Bash completion is shipped as a generated file in [`completions/kpx.bash`](./completions/kpx.bash).

Install it on macOS with Homebrew Bash completion:

```bash
brew install bash-completion@2
mkdir -p "$(brew --prefix)/etc/bash_completion.d"
cp ./completions/kpx.bash "$(brew --prefix)/etc/bash_completion.d/kpx"
```

Install it on Linux for the current shell:

```bash
mkdir -p ~/.local/share/bash-completion/completions
cp ./completions/kpx.bash ~/.local/share/bash-completion/completions/kpx
```

If you use release archives, extract the matching macOS or Linux tarball and place `kpx` somewhere on your `PATH`, for example `~/.local/bin`.

## Quick Start

Create a vault:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin db create ./vault.kdbx --name "Personal Vault"
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

Print only the password for piping to `pbcopy`:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin entry password ./vault.kdbx /Personal/GitHub | pbcopy
```

Search by title:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin find ./vault.kdbx github
```

Create a paper backup:

```bash
printf '%s\n' 'master-password' | ./kpx --master-password-stdin export paper ./vault.kdbx \
  --output ./vault-paper-backup.txt
```

## Configuration

Create `~/.kpx/config.yml`:

```yaml
default_database: /Users/you/vault.kdbx
reveal: false
master_password_cache_seconds: 0
backup_directory: ""
backup_filename_format: "{db_filename}.{timestamp}.{db_ext}"
save_method: "temporary_file"
```

Set `master_password_cache_seconds` to a positive number to cache the master password for that many seconds. The default is `0`, which disables caching.

Leave `backup_directory` empty to store backups alongside the database. The default filename format uses the original database filename plus a UTC timestamp.

`save_method` defaults to `"temporary_file"`. Set it to `"direct_write"` to write directly to the database file instead.

Available backup filename placeholders:

- `{db_filename}`: database filename without the extension
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
kpx export ...
kpx find ...
```

Version checks:

```bash
kpx version
kpx --version
```

JSON output:

```bash
kpx --json entry show ./vault.kdbx /Personal/GitHub
```

JSON support is available for the commands listed below, but the output schema may still evolve while the project is in early releases.

Available today:

- `kpx db create`
- `kpx db validate`
- `kpx group ls`
- `kpx group add`
- `kpx entry ls`
- `kpx entry show`
- `kpx entry password`
- `kpx entry add`
- `kpx entry edit`
- `kpx entry rm`
- `kpx export paper`
- `kpx find`
- `kpx version`
- `kpx --version`

Path rules:

- group paths look like `/Personal/Email`
- entry paths look like `/Personal/GitHub`
- the database argument is optional when `~/.kpx/config.yml` defines `default_database`
- `entry show` uses `reveal` from config unless `--reveal` is explicitly passed
- `--json` emits machine-readable output for supported commands, but the JSON schema is not guaranteed stable yet
- `--master-password-stdin` is the standard flag for reading the database master password from stdin
- `--entry-password-stdin` is the standard flag for reading an entry password from stdin
- `--no-input` disables interactive prompts and requires stdin flags or cached credentials for secret input
- `entry rm` requires `--force` when `--no-input` is set
- ambiguous matches fail closed

Non-interactive examples:

```bash
printf '%s\n' 'master-password' | ./kpx --no-input --master-password-stdin db validate ./vault.kdbx
printf '%s\n' 'master-password' | ./kpx --no-input --master-password-stdin entry rm ./vault.kdbx /Personal/GitHub --force
printf '%s\n%s\n' 'master-password' 'entry-password' | ./kpx --no-input --master-password-stdin entry add ./vault.kdbx /Personal/GitLab --entry-password-stdin
```

## Security Notes

- secrets are redacted by default
- password prompts use the controlling tty and do not echo
- master password caching is disabled by default
- when enabled, cached master passwords are stored on disk under `~/.kpx/master-password-cache.yml` with restrictive file permissions and a cooperating file lock
- cooperating `kpx` processes use a stable adjacent lock file with restrictive permissions to coordinate reads and writes
- database saves create a backup of the existing file before replacing it
- save method defaults to temporary-file-then-rename and can be changed to direct write in config
- cooperating `kpx` processes take advisory locks during database reads and hold an exclusive lock across write operations
- non-interactive usage supports stdin-based secrets
- writes are atomic when `save_method` is left at the default `temporary_file`
- destructive entry deletion requires confirmation unless `--force` is provided

Current limitations:

- key files are not supported yet
- advisory locks coordinate cooperating `kpx` processes, not every external KeePass tool
- `direct_write` is available for compatibility, but `temporary_file` remains the safer default
- paper export writes plaintext secrets and should be handled like a physical recovery artifact

## Troubleshooting

- `no interactive tty available`
  Use `--master-password-stdin` for database passwords and `--entry-password-stdin` for entry passwords when running non-interactively.
- `interactive input disabled`
  `--no-input` turns off prompts entirely. Pair it with stdin flags or a configured master-password cache.
- `delete requires --force when --no-input is set`
  Add `--force` to `entry rm` when scripting deletes.
- advisory locking unsupported on this platform
  Real vault operations are currently supported on Unix-like systems only.
- release archive command not found after extraction
  Move the extracted `kpx` binary into a directory on your `PATH`, such as `~/.local/bin`.

## Roadmap

Planned next:

- key file support
- password + key file support
- group rename, move, and delete
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

Create a release:

```bash
# update internal/buildinfo/VERSION.txt and any user-facing version references first
git tag v$(tr -d '\n' < internal/buildinfo/VERSION.txt)
git push origin main --tags
```

Pushing a `v*` tag runs the GitHub release workflow, verifies the tag matches `internal/buildinfo/VERSION.txt`, runs `go test ./...`, and publishes release archives plus checksums.
Release archives are currently published for macOS and Linux targets only. Windows release packaging is out of scope for now.

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

- desktop GUI
- browser integration
- desktop integration
- SSH agent integration
- auto-type
- sync and merge
- attachments and entry history
- `KDBX3` support

## License

MIT. See [LICENSE](./LICENSE).
