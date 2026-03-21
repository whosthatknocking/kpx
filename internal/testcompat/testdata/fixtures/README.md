# Compatibility Fixtures

This directory holds compatibility fixtures for `kpx`.

## Current state

The automated suite currently includes one `generated` baseline fixture through
[`manifest.json`](./manifest.json). That gives us immediate coverage for:

- open and decrypt
- inspect groups and entries
- title search
- save and reopen without losing protected values

## Adding real KeePassXC fixtures

To expand this into true cross-tool compatibility coverage:

1. Create a `.kdbx` fixture with KeePassXC.
2. Copy it into this directory, for example `keepassxc-aes-argon2id.kdbx`.
3. Add a `source: "file"` entry to [`manifest.json`](./manifest.json).
4. Include:
   - `path`
   - `password`
   - expected `groups`
   - expected `entries`
   - any expected `searches`

The test harness copies file fixtures into a temp directory before saving, so
the original checked-in fixture stays unchanged.

## Recommended next fixtures

- KeePassXC AES-256 + Argon2id
- KeePassXC AES-256 + AES-KDF
- KeePassXC ChaCha20 + Argon2id
- fixtures with custom fields
