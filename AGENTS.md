# AGENTS.md

This file gives project-specific guidance to AI agents working in this repository.

## Project Context

- Project: `kpx`
- Purpose: provide a focused, scriptable Go CLI for working with KeePassXC-compatible `KDBX4` password databases
- Runtime: Go `1.25+`
- Product shape:
  - CLI only
  - `KDBX4` only
  - macOS-first, with Unix-style advisory locking for cooperating `kpx` processes
- Primary concerns:
  - safe reads and writes
  - predictable scripting behavior
  - secrets redacted by default
  - compatibility with KeePassXC workflows

## Source of Truth

When behavior, naming, or scope is unclear, use these files in this order:

1. `docs/PROJECT_SPEC.md`
2. `README.md`
3. command implementations under `cmd/`
4. vault and storage implementations under `internal/`

Keep docs aligned with implementation. If you change command names, flags, config behavior, output shapes, save behavior, or security-sensitive behavior, update the relevant docs in the same task.

## Architecture Map

- `main.go`
  - process entrypoint
  - maps returned errors to CLI exit codes and stderr output
- `cmd/root.go`
  - Cobra root command
  - global flags such as `--json`, `--no-input`, and `--master-password-stdin`
  - version output wiring
- `cmd/db.go`
  - database creation and validation commands
- `cmd/group.go`
  - group listing and creation commands
- `cmd/entry.go`
  - entry list/show/password/add/edit/remove command wiring
- `cmd/find.go`
  - title and path search command wiring
- `cmd/export.go`
  - plaintext paper export flow
- `cmd/output.go`
  - human-readable and JSON output helpers
- `internal/vault/`
  - database open/save lifecycle
  - group and entry path resolution
  - entry and group mutations
  - `gokeepasslib` integration
- `internal/store/`
  - atomic writes
  - direct-write fallback
  - backup creation
  - advisory file locking on Unix-like systems
  - explicit unsupported-platform stub for non-Unix builds
- `internal/config/`
  - `~/.kpx/config.yml` load/save behavior
- `internal/cache/`
  - master password cache persistence and expiry
- `internal/cli/`
  - secret input, confirmation prompts, field parsing, and exit-error helpers
- `internal/export/`
  - plaintext paper export rendering
- `internal/buildinfo/`
  - base version source and version formatting
- `internal/testcompat/`
  - compatibility tests and KeePassXC fixture handling
- `tools/gen_bash_completion.go`
  - completion generation source for `completions/kpx.bash`

## Non-Negotiable Design Rules

- Keep the project CLI-first. Do not add GUI, browser, or daemon-style behavior casually.
- Preserve `KDBX4` and KeePassXC compatibility.
- Favor secure defaults over convenience:
  - secrets redacted by default
  - non-interactive secret handling must be explicit
  - destructive actions should fail closed unless the user opts in clearly
- Keep writes safe:
  - preserve advisory locking behavior for cooperating `kpx` processes
  - keep backup-before-save behavior intact unless the change explicitly redesigns it
  - prefer atomic saves unless there is a documented reason not to
- Maintain scriptability. Human-readable output and JSON output should remain predictable.
- Avoid broadening scope with unrelated password-manager features that are explicitly out of scope in `docs/PROJECT_SPEC.md`.

## Command and Package Conventions

- Add or change CLI surface area under `cmd/`.
- Keep command help text concise and precise; it is user-facing documentation.
- Route KDBX operations through `internal/vault` rather than duplicating database logic in command handlers.
- Route file persistence, backup, and locking concerns through `internal/store`.
- Route config reads and writes through `internal/config`.
- Route password prompting, stdin secret reads, confirmations, and structured CLI exit errors through `internal/cli`.
- Keep JSON payloads stable within the current command design. If you change a JSON response shape, update tests and docs in the same task.
- Keep entry and group path handling conservative. Do not silently reinterpret ambiguous paths.
- Preserve the distinction between redacted default output and explicit secret reveal flows.

## Security and Stability Expectations

- Treat plaintext secrets, password cache behavior, paper export output, and save paths as security-sensitive areas.
- Do not print secrets in normal success messages, logs, or errors unless a command is explicitly designed to reveal them.
- Do not weaken file permissions for config, cache, temp files, backups, or database writes without a strong reason.
- If a change affects `--master-password-stdin`, `--entry-password-stdin`, caching, or paper export, review the non-interactive and leakage implications carefully.
- Preserve clear exit-error behavior for authentication failures, format problems, missing paths, ambiguous matches, and save failures.
- Advisory locks coordinate cooperating `kpx` processes only; do not overstate that as universal cross-tool locking.

## Testing Expectations

Run the smallest relevant test set first, then broaden if needed.

- Main suite: `go test ./...`
- Common targeted runs:
  - `go test ./cmd/...`
  - `go test ./internal/vault/... ./internal/store/... ./internal/export/...`
  - `go test ./internal/testcompat/...`

Testing guidance:

- Add or update tests for any behavior change in command parsing, output, path handling, save behavior, backup behavior, locking, config behavior, cache behavior, or export formatting.
- Prefer focused package tests first when iterating.
- Run the full `go test ./...` suite before finishing when you changed behavior.
- Keep CI expectations aligned with the release surface:
  - Linux runs the main build, vet, test, and release-target cross-build checks
  - macOS runs build and test coverage in CI
- If you change command completions, regenerate and verify `completions/kpx.bash`.
- If you could not run some validation, say so explicitly in your summary.

## Documentation Expectations

Update docs when any of these change:

- command names or flags
- JSON output shapes
- configuration keys or behavior
- password input behavior
- backup, locking, or save semantics
- platform support statements
- install or build instructions

Common files to update:

- `README.md`
- `docs/PROJECT_SPEC.md`
- `completions/kpx.bash` when CLI completion output changes

## Practical Workflow

1. Read the affected command and the matching lower-level package first.
2. Make the smallest coherent change.
3. Update tests alongside the behavior change.
4. Update docs if user-facing behavior changed.
5. Run targeted tests, then `go test ./...` when appropriate.

## Commit and PR Guidance

- Use imperative commit subjects, for example `cmd: clarify paper export stdout requirement`.
- Keep commits focused and easy to review.
- Include tests with behavior changes when practical.
- Avoid mixing unrelated refactors with behavior or docs changes unless they are tightly coupled.
- In PRs, summarize the user-visible change and call out validation actually run.
- If validation was limited, say so explicitly.

## Repository-Specific Notes

- The base release version is stored in `internal/buildinfo/VERSION.txt`.
- The generated bash completion file lives at `completions/kpx.bash`.
- The config file is `~/.kpx/config.yml`.
- The master password cache file is `~/.kpx/master-password-cache.yml`.
- Non-Unix builds compile with a clear unsupported-platform lock stub; real vault operations are currently intended for Unix-like systems only.
- Current implemented CLI areas include:
  - `db create`
  - `db validate`
  - `group ls`
  - `group add`
  - `entry ls`
  - `entry show`
  - `entry password`
  - `entry add`
  - `entry edit`
  - `entry rm`
  - `find`
  - `export paper`
  - `version`
  - `--version`
- The project intentionally remains narrower than a full desktop password manager. Features like GUI work, browser integration, key files, merge/sync, and broad importer/exporter ecosystems are out of scope or not implemented yet.

## Good Changes

- tightening path or output handling with tests
- improving save, backup, or lock robustness without weakening safety guarantees
- clarifying JSON output behavior and documenting it
- improving KeePassXC compatibility coverage
- fixing redaction, stdin secret handling, or confirmation edge cases
- keeping command code thin while moving reusable logic into the right internal package

## Bad Changes

- adding GUI or desktop-integration features to the CLI project without explicit direction
- printing secrets by default or weakening reveal safeguards
- bypassing `internal/vault` or `internal/store` for database mutations and persistence
- changing JSON output or config behavior without updating tests and docs
- weakening file permissions or lock behavior casually
- introducing broad new product scope that conflicts with `docs/PROJECT_SPEC.md`
