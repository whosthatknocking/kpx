# kpx

`kpx` is a focused Go CLI for working with KeePassXC-compatible `KDBX4` password databases.

It aims to be small, scriptable, and easy to audit:

- CLI only
- `KDBX4` only
- macOS-first
- safe atomic saves
- secrets redacted by default

## Why

`kpx` is intentionally narrower than a full KeePass desktop replacement. The goal is a dependable command-line tool for people who want to:

- create and maintain local encrypted password databases
- script common vault operations
- inspect and update entries without leaving the terminal
- keep compatibility with KeePassXC workflows

## Current Status

The project has moved beyond the spec-only stage and now includes a working MVP codebase with tests.

Implemented today:

- create a new database
- validate and open an existing database
- list and create groups
- list, show, add, edit, and delete entries
- search entries by title
- show the CLI version via `kpx version` and `kpx --version`
- atomic save behavior
- a refactored package layout with smaller command and vault files
- round-trip and fixture-based test coverage

The detailed product plan still lives in [PROJECT_SPEC.md](./PROJECT_SPEC.md).

## Features

### MVP

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

### Planned Next

- key file support
- password + key file support
- group rename, move, and delete
- JSON output
- stronger KeePassXC fixture coverage

## Install

Build from source:

```bash
git clone https://github.com/whosthatknocking/kpx.git
cd kpx
go build ./...
```

Build the CLI binary:

```bash
go build -o kpx .
```

To embed a release version in the binary:

```bash
go build -ldflags "-X github.com/whosthatknocking/kpx/internal/buildinfo.Version=v0.1.0" -o kpx .
```

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

## Command Shape

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

Path rules:

- group paths look like `/Personal/Email`
- entry paths look like `/Personal/GitHub`
- ambiguous matches fail closed

## Security Notes

- secrets are redacted by default
- password prompts use the controlling tty and do not echo
- non-interactive usage supports stdin-based secrets
- writes are atomic
- destructive entry deletion requires confirmation unless `--force` is provided

## Compatibility

`kpx` is currently built around [`tobischo/gokeepasslib/v3`](https://pkg.go.dev/github.com/tobischo/gokeepasslib/v3) behind an internal adapter.

The test suite currently includes:

- CLI end-to-end flow tests
- repeated save/open round-trip tests
- atomic write tests
- fixture-driven compatibility tests

Fixture infrastructure for real KeePassXC-generated `.kdbx` files is ready in [internal/testcompat/testdata/fixtures](./internal/testcompat/testdata/fixtures), and the next compatibility step is to add more real fixture files covering multiple KDF and cipher combinations.

## Development

Run the test suite:

```bash
go test ./...
```

Project layout:

```text
cmd/                Cobra command definitions
internal/buildinfo/ version metadata exposed to the CLI
internal/cli/       prompting, exit handling, and CLI helpers
internal/store/     atomic file writes
internal/testcompat/fixture-driven compatibility tests
internal/vault/     KDBX-backed vault adapter
```

## Roadmap

1. Expand KeePassXC fixture coverage with real checked-in databases
2. Add key file support
3. Add JSON output for scripting
4. Add group rename, move, and delete
5. Harden compatibility and release packaging

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

No license file has been added yet.
