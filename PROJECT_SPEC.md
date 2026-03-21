# KeePass-Compatible Go CLI Project Spec

## 1. Summary

Build a macOS-first CLI password manager in Go with a deliberately minimal first milestone. The project should start with the smallest feature set that is still genuinely useful and KeePassXC-compatible, then expand carefully:

- CLI only
- `KDBX4` databases only
- KeePassXC-compatible database files
- No browser integration
- No desktop GUI integration
- No SSH agent, auto-type, or secret-service integration in v1

Working project name: `kpx`

The tool should be useful as both:

- a standalone password manager for terminal users
- a scriptable automation-friendly CLI for CI, local scripting, and shell workflows

## Implementation Status

Current repo status as of March 21, 2026:

- [x] Git repository initialized
- [x] GitHub remote configured (`origin`)
- [x] Initial project spec written
- [x] Initial commit created and pushed
- [x] Dedicated `mvp` branch created
- [x] `mvp` merged into `main`
- [x] Go module initialized
- [x] CLI project skeleton created
- [x] KDBX integration started
- [x] Core CLI and vault code split into smaller files by concern
- [x] Automated test suite currently passes with `go test ./...`
- [x] Current Go implementation committed to Git
- [x] MIT license added
- [x] README prepared for GitHub publication

Current product status in the workspace:

- [x] `db create`
- [x] `db validate`
- [x] `group ls`
- [x] `group add`
- [x] `entry ls`
- [x] `entry show`
- [x] `entry add`
- [x] `entry edit`
- [x] `entry rm`
- [x] `find`
- [x] `version`
- [x] `--version`
- [x] atomic save support
- [x] optional default database config in `~/.kpx/config.yml`
- [x] optional default reveal behavior in `~/.kpx/config.yml`
- [x] optional master password cache duration in `~/.kpx/config.yml`
- [x] database backup before save with configurable destination and filename format
- [x] configurable save method with temporary-file writes as the default
- [x] centralized release version source in `internal/buildinfo/VERSION.txt`
- [x] official on-demand KeePassXC fixtures pinned to upstream commit and SHA-256
- [x] advisory file locking for cooperating `kpx` processes
- [x] secure password prompt support
- [x] paper export for printed emergency recovery
- [x] JSON output for supported commands
- [ ] key file support
- [ ] group rename/move/delete

## 2. Goals

### Primary goals

- [x] Read and write KeePassXC-compatible `KDBX4` databases.
- [x] Ship a small, finishable MVP before adding broader parity features.
- [x] Keep the codebase small, maintainable, and easy to audit.
- [x] Favor secure defaults and predictable non-interactive behavior.
- [x] Work on macOS.

### Product goals

- [x] Fast startup and low operational complexity.
- [x] Human-friendly interactive use.
- [ ] Stable machine-readable output modes for scripting.
- [x] Minimal external dependencies unless they clearly reduce complexity and risk.

## 3. Scope Levels

### MVP

The minimum viable product should support only:

- [x] create a new `KDBX4` database
- [x] open an existing `KDBX4` database
- [x] save changes safely
- [x] list groups
- [x] create groups
- [x] list entries
- [x] show entries
- [x] add entries
- [x] edit entries
- [x] delete entries
- [x] search entries by title

### v1

After MVP is stable, v1 can add:

- group rename/move/delete
- key file support
- password + key file combination
- credential change command
- stronger compatibility and round-trip coverage

### Later

Explicitly deferred until after v1:

- attachments
- entry history
- advanced search across multiple fields
- interactive shell/REPL
- clipboard support
- reports

## 4. Non-Goals

The following are explicitly out of scope for v1:

- Desktop GUI
- Browser extension integration
- Native desktop integration
- Auto-type
- SSH agent integration
- Passkeys / WebAuthn
- YubiKey / challenge-response
- Database merge / sync features
- Importers from Bitwarden, 1Password, CSV, etc.
- XML/HTML/CSV export beyond narrow debugging/admin use
- KDBX 3.x or KDB support
- Plugin ecosystem

## 5. Target Users

- Developers who want terminal-native secret management.
- Existing KeePassXC users who want a scriptable companion tool.
- Ops / SRE / platform engineers who need shell automation around a KeePass database.
- Privacy-focused users who want local encrypted storage without cloud coupling.

## 6. Success Criteria

The project is successful when a user can do all of the following entirely with the CLI:

- create a new `KDBX4` database
- open existing KeePassXC-created `KDBX4` databases
- list groups and entries
- inspect, add, edit, and delete entries
- create groups
- search entries by title
- save the database so KeePassXC can open it without warnings or corruption

## 7. Functional Scope

### 7.1 MVP database operations

Required:

- [x] create database
- [x] open database
- [x] validate/decrypt database
- [x] save database
- [x] save atomically
- [x] back up existing database before save
- [x] configurable save method
- [x] prompt for master password securely

CLI examples:

```bash
kpx db create vault.kdbx
kpx db validate vault.kdbx
```

### 7.2 MVP group operations

Required:

- [x] list groups as paths
- [x] create group

CLI examples:

```bash
kpx group ls vault.kdbx
kpx group add vault.kdbx /Personal/Email
```

### 7.3 MVP entry operations

Required:

- [x] list entries
- [x] show entry details
- [x] add entry
- [x] edit entry fields
- [x] delete entry

Core built-in fields:

- `Title`
- `UserName`
- `Password`
- `URL`
- `Notes`

Custom fields:

- arbitrary string fields
- protected and unprotected values where format allows

CLI examples:

```bash
kpx entry ls vault.kdbx /Personal
kpx entry show vault.kdbx /Personal/GitHub
kpx entry add vault.kdbx /Personal/GitHub --username alice
kpx entry edit vault.kdbx /Personal/GitHub --url https://github.com
kpx entry rm vault.kdbx /Personal/GitHub
```

### 7.4 MVP search

Required:

- [x] search by title
- [x] exact and substring matching

CLI examples:

```bash
kpx find vault.kdbx github
```

### 7.5 v1 additions

Planned additions after MVP:

- group rename/move/delete
- key file support
- password + key file combination
- credential change command

### 7.5.1 Paper export for emergency recovery

Goal:

- export a database into a human-readable plaintext format intended for secure
  printing and physical storage

This export is for readable secret recovery, not machine interchange.

Required behavior:

- include a document header once at the top
- include export timestamp once at the top
- include tool version once at the top
- include database display name once at the top when available
- include source database path once at the top
- include entry path, title, username, password, URL, notes, and custom fields
- omit low-level database internals and implementation-heavy metadata
- use stable ordering for groups, entries, and custom fields
- be easy to read, print, and copy/paste

Must exclude by default:

- UUIDs
- raw XML
- KDBX header details
- cryptographic metadata
- internal timestamps
- deleted object metadata
- history
- attachments

Safety expectations:

- the command must warn that plaintext secrets are being exported
- the command should require confirmation unless `--force` is provided
- it should prefer file output
- it should only write plaintext secrets to stdout when explicitly requested

Recommended command shape:

```bash
kpx export paper <database>
kpx export paper <database> --output vault-paper-backup.txt
```

Example header:

```text
kpx Paper Backup
Generated: 2026-03-21T18:42:00Z
Tool Version: 0.1.6
Database: Personal Vault
Source File: /Users/you/vault.kdbx
```

Example entry block:

```text
========================================================================
Path: /Personal/GitHub
Title: GitHub
UserName: alice
Password: super-secret
URL: https://github.com
Notes: Personal account

Custom Fields:
Environment: prod
Recovery Code: ABCD-EFGH-IJKL
```

### 7.6 Later additions

Deferred until after v1:

- attachments
- entry history
- advanced search fields
- interactive shell/REPL
- clipboard support
- report commands

### 7.7 Output modes

Required:

- pretty human-readable output
- `--quiet` mode
- version output via command and flag
- stable exit codes

Planned additional output mode:

- paper-oriented plaintext export for emergency recovery
- JSON output via `--json` for supported commands

Current note:

- JSON output is implemented for supported commands, but the schema may still evolve until a stable contract is declared.

### 7.8 Optional user config

Supported:

- optional config file in `~/.kpx/config.yml`
- optional `default_database` setting
- optional `reveal` setting for `entry show`
- optional `master_password_cache_seconds` setting, defaulting to disabled
- optional `backup_directory` and `backup_filename_format` settings for pre-save backups
- optional `save_method` setting with `temporary_file` as the default and `direct_write` as the alternative
- commands that open an existing vault may omit the database path when `default_database` is configured
- explicit command-line database arguments override config
- explicit CLI `--reveal` overrides config
- config management is file-based only; there are no `kpx config ...` subcommands

### 7.9 Example workflows

Create a new vault:

```bash
kpx db create ~/vault.kdbx
```

Create a group and add an entry with a password from stdin:

```bash
kpx group add ~/vault.kdbx /Personal
printf '%s' 'my-secret-password' | kpx entry add ~/vault.kdbx /Personal/GitHub \
  --username alice \
  --url https://github.com \
  --notes 'Personal account' \
  --master-password-stdin
```

List entries and inspect one entry:

```bash
kpx entry ls ~/vault.kdbx /Personal
kpx entry show ~/vault.kdbx /Personal/GitHub
```

Search for an entry and update its URL:

```bash
kpx find ~/vault.kdbx github
kpx entry edit ~/vault.kdbx /Personal/GitHub --url https://github.com/login
```

Generate a printable emergency recovery export:

```bash
kpx export paper ~/vault.kdbx --output ~/vault-paper-backup.txt
```

## 8. Feature Priorities

### Phase 1: Must-have MVP

- [x] database create/open/save
- [x] master password support
- [x] group list/create
- [x] entry CRUD
- [x] title search
- [x] atomic saves
- [x] KeePassXC compatibility validation tests
- [x] backup before save
- [x] optional default database config
- [x] version command and build version metadata

### Phase 2: Strongly desired for v1.0

- key file support
- group rename/move/delete
- credential change command

### Phase 3: Nice-to-have after v1.0

- attachments
- entry history
- interactive REPL/shell mode
- shell completion
- manpage generation
- report commands
- clipboard support

## 9. Compatibility Requirements

### 8.1 File format

The tool must support `KDBX4` only.

It must:

- open KeePassXC-created `KDBX4` databases
- save databases that KeePassXC can reopen successfully
- preserve data it does not semantically understand whenever possible

### 8.2 Required crypto compatibility

To be realistically compatible with modern KeePassXC usage, support should include:

- outer ciphers:
  - AES-256-CBC
  - ChaCha20
- KDFs:
  - Argon2id as preferred default for new databases
  - Argon2d for compatibility
  - AES-KDF for opening older-but-still-KDBX4 databases
- compression:
  - none
  - gzip
- protected inner fields:
  - KDBX4 inner stream handling compatible with KeePassXC/KeePass conventions

Notes:

- KeePassXC documents `AES-KDF` and `Argon2` support for `KDBX4`, and recommends `Argon2id`.
- KeePass introduced `ChaCha20` in `KDBX4`.
- Avoid `Twofish` in v1 unless required by the chosen Go KDBX library and proven interoperable. It is not necessary for the initial "simple but practical" target.

### 9.3 Credential support

Required for MVP:

- master password

Required for v1:

- key file
- password + key file combination

Out of scope:

- Windows account binding
- hardware-backed challenge response

### 9.4 Data fidelity

MVP must preserve:

- UUIDs
- timestamps
- entry field data
- group structure

v1 should additionally preserve when supported by the library:

- custom icons
- deleted objects metadata
- custom data blobs

## 10. Security Requirements

### 9.1 Secure defaults

- New databases default to `KDBX4 + Argon2id`.
- New databases default to `AES-256-CBC` unless ChaCha20 is explicitly requested.
- Writes must be atomic: write temp file, fsync, rename.
- Writes should create a backup of the existing database before replacement.
- Never print secrets unless explicitly requested.
- Password prompts must not echo.
- Sensitive fields should be zeroed or dropped from memory as much as practical in Go.

### 9.2 CLI behavior

- `show` commands must redact passwords by default.
- `--reveal` must be explicit.
- `--json` output must still redact secrets unless `--reveal` is also passed.
- JSON support may expand over time, and its schema should not be treated as frozen until the project declares it stable.
- Commands that can destroy data must require confirmation unless `--force` is provided.

### 9.3 Process and shell safety

- Avoid putting secrets in process arguments when interactive prompting is possible.
- Support reading secrets from stdin for automation.
- Provide `--master-password-stdin` for database passwords and `--entry-password-stdin` for entry passwords in non-interactive flows.
- Document shell history risks clearly.
- Keep master password caching disabled by default.

### 9.4 File safety

- Respect file permissions on save.
- Warn when database file is world-readable.
- Use advisory file locking when feasible.

## 11. UX Requirements

### 10.1 General CLI design

- Subcommand-oriented interface.
- Predictable noun/verb structure.
- Consistent flags across commands.
- Helpful errors with next-step hints.

Recommended command shape:

```bash
kpx db ...
kpx group ...
kpx entry ...
kpx find ...
```

### 10.2 Path addressing

Entries and groups should be addressable by:

- full group path
- UUID
- exact title within a group

If a path is ambiguous, the tool must fail with a clear disambiguation message.

Database selection should support:

- explicit database path argument
- optional default database from user config for commands that open an existing vault
- predictable precedence where explicit CLI arguments win over config

### 11.3 Scripting behavior

- No interactive prompts when `--no-input` is set
- Clear non-zero exit codes for:
  - authentication failure
  - file format error
  - not found
  - ambiguous match
  - validation/save failure

## 12. Suggested Command Set

### Database

- `kpx db create`
- `kpx db validate`

### Groups

- `kpx group ls`
- `kpx group add`

### Entries

- `kpx entry ls`
- `kpx entry show`
- `kpx entry add`
- `kpx entry edit`
- `kpx entry rm`

### Search

- `kpx find`

### Config

- optional `~/.kpx/config.yml`
- `default_database` setting

### Metadata

- `kpx version`
- `kpx --version`

### v1 additions

- `kpx db change-master`
- `kpx group mv`
- `kpx group rm`

## 13. Architecture

### 13.1 High-level modules

- `cmd/`
  - Cobra command definitions
- `internal/cli/`
  - flag parsing helpers, prompt helpers, output formatting
- `internal/config/`
  - optional user config loading and saving
- `internal/buildinfo/`
  - embedded version metadata exposed to the CLI
- `internal/cache/`
  - optional master password cache handling
- `internal/store/`
  - file IO, atomic writes, lock handling
- `internal/testcompat/`
  - compatibility fixtures and round-trip tests
- `internal/vault/`
  - KDBX-backed vault adapter and path-based operations

### 13.2 Design principles

- Keep the KDBX library behind an adapter boundary.
- Minimize direct library-specific types outside `internal/kdbx/`.
- Prefer pure functions and explicit inputs/outputs.
- Keep business logic separate from CLI rendering.
- Keep files focused by concern so command wiring, output helpers, and vault operations stay easy to review.
- Make it easy to swap KDBX libraries if the initial choice proves insufficient.

## 14. Dependency Strategy

Use external libraries only where they materially reduce code and maintenance burden.

### 14.1 Recommended dependencies

- CLI framework: [`github.com/spf13/cobra`](https://github.com/spf13/cobra)
  - Widely used, actively maintained, excellent fit for subcommand CLIs.
- Terminal password input: `golang.org/x/term`
  - Official Go subrepository; avoids fragile prompt handling.
- KDBX library: prefer [`github.com/tobischo/gokeepasslib/v3`](https://pkg.go.dev/github.com/tobischo/gokeepasslib/v3)
  - Current tagged releases, stable Go module metadata, and explicit `KDBX4` APIs.

### 14.2 KDBX library decision

Current recommendation: start with `gokeepasslib/v3`.

Reasoning:

- It has a stable tagged Go module.
- Public docs expose `KDBX4`-specific constructors and encoder/decoder paths.
- It appears substantially more mature than smaller alternatives like `cixtor/kdbx`.

Risk:

- Real interoperability must be proven with fixture-based tests against KeePassXC-created databases.
- If write compatibility or fidelity is insufficient, replace the adapter implementation without changing CLI-facing code.

### 14.3 Dependencies to avoid unless needed

- Full TUI frameworks
- Clipboard libraries
- ORM/config-heavy frameworks
- Reflection-heavy CLI abstractions

## 15. Data Model Expectations

### 15.1 Entry fields

The app-level entry abstraction should include:

- UUID
- group path
- title
- username
- password
- URL
- notes
- custom fields
- timestamps

### 15.2 Group fields

- UUID
- path
- name
- notes if supported by underlying data model
- timestamps
- children

## 16. Testing Strategy

### 16.1 Unit tests

Cover:

- path parsing
- search matching
- output rendering
- paper export rendering and field omission rules

### 16.2 Compatibility tests

Maintain a fixture set of `KDBX4` databases created by KeePassXC, including combinations of:

- AES-256 + Argon2id
- AES-256 + AES-KDF
- ChaCha20 + Argon2id
- password only
- keyfile only for v1
- password + keyfile for v1
- with custom fields

Required compatibility checks:

- open fixture successfully
- verify expected metadata and entry content
- save without losing fields
- support official on-demand upstream fixtures pinned by commit and checksum
- reopen in KeePassXC manually or through compatibility validation workflow

### 16.3 Round-trip tests

- create with `kpx`, open with KeePassXC
- create with KeePassXC, modify with `kpx`, reopen with KeePassXC
- repeated save cycles do not corrupt data

### 16.4 Security tests

- no password echo in prompts
- temp files removed on error
- atomic save rollback behavior
- paper export confirmation flow for plaintext secret output

## 17. Release Criteria

### MVP release criteria

All of the following must be true for the first usable release:

- [x] supports `KDBX4` open/save for real KeePassXC fixture databases
- [x] supports group list/create
- [x] supports entry CRUD
- [x] supports title search
- [x] supports atomic save behavior
- [x] supports master-password-based unlock
- [x] test suite passes on macOS
- [x] documentation includes quickstart and threat notes

### v1 release criteria

- supports key files
- supports group rename/move/delete
- expands compatibility coverage beyond MVP fixtures

## 18. Documentation Requirements

Required docs:

- README with install and quickstart
- security model and limitations
- scripting examples
- compatibility notes with KeePassXC
- optional config file behavior and default database workflow
- version command and release-version build notes
- command reference

## 19. Risks and Mitigations

### Risk: Go KDBX libraries may not fully preserve KeePassXC semantics

Mitigation:

- isolate KDBX implementation behind adapter layer
- build fixture-heavy compatibility tests early
- validate write fidelity before feature expansion

### Risk: Secret handling in Go is imperfect

Mitigation:

- minimize secret lifetime in memory
- avoid logging sensitive values
- never expose secrets by default
- document limitations honestly

### Risk: Feature creep from early parity ambitions

Mitigation:

- define a strict MVP before v1
- treat browser, GUI, SSH, auto-type, sync, import/export, attachments, and history as out of scope for MVP

## 20. Recommended Default Choices

For new databases:

- format: `KDBX4`
- KDF: `Argon2id`
- cipher: `AES-256-CBC`
- compression: `gzip`

For CLI behavior:

- redact secrets by default
- fail closed on ambiguity

## 21. Proposed Milestones

### Milestone 1: Foundation

- [x] project skeleton
- [x] command layout
- [x] database open/save adapter
- [x] password prompt flow
- [x] fixture loading tests

### Milestone 2: Core Vault Ops

- [x] group list/create
- [x] entry CRUD
- [x] title search

### Milestone 3: v1 Expansion

- [x] paper export for printed emergency recovery
- [ ] key file support
- [ ] group rename/move/delete

### Milestone 4: Post-v1 Features

- attachments
- history

### Milestone 5: Hardening

- [x] atomic saves
- [x] locking
- [x] compatibility matrix
- [x] docs and release packaging

### Milestone 6: Portability

- Linux support
- broader cross-platform packaging

## 22. Final Recommendation

Build `kpx` as a focused Go CLI with a strict MVP, not as a full KeePassXC clone.

The best MVP interpretation is:

- local database create/open/save
- group list/create
- entry CRUD
- title search

The best v1 expansion after that is:

- key file support
- group move/delete

while explicitly excluding in the early phases:

- browser integration
- desktop integrations
- advanced ecosystem features
- attachments
- history

That keeps the first release small enough to finish while still delivering a genuinely useful, KeePassXC-compatible tool.
