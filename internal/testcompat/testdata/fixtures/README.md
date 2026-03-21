# Compatibility Fixtures

This directory holds compatibility fixtures for `kpx`.

## Current state

The automated suite currently includes one `generated` baseline fixture through
[`manifest.json`](./manifest.json). That gives us immediate coverage for:

- open and decrypt
- inspect groups and entries
- title search
- repeated save and reopen cycles without losing protected values
- both configured save methods
- backup file reopen checks after save

Current limitation:

- there are still no checked-in real KeePassXC-generated `.kdbx` fixtures in this repository
- live validation against a locally installed `keepassxc` or `keepassxc-cli` is still separate from these tests

## Remote fixtures

The manifest supports fetching real `.kdbx` fixtures on demand instead of
checking them into this repository. Remote fixtures are opt-in so everyday test
runs stay offline and deterministic.

Run them with:

```bash
KPX_REMOTE_FIXTURES=1 go test ./internal/testcompat -run TestCompatibilityFixtures
```

Use a manifest entry like:

```json
{
  "name": "keepassxc-real-aes-argon2id",
  "source": "url",
  "url": "https://example.invalid/fixture.kdbx",
  "sha256": "replace-with-expected-sha256",
  "password": "hunter2",
  "database_name": "Fixture",
  "groups": ["/Personal"],
  "entries": [
    {
      "path": "/Personal/GitHub",
      "username": "alice",
      "password": "super-secret",
      "url": "https://github.com",
      "notes": "Personal account"
    }
  ]
}
```

Notes:

- prefer pinned URLs to exact release assets or exact raw file revisions
- prefer providing `sha256` so fixture contents are verified after download
- avoid using floating `latest` URLs in automated tests because they reduce reproducibility

## Official KeePassXC fixtures in use

The manifest currently includes official fixtures from the KeePassXC upstream
repository pinned to a specific upstream commit:

- `tests/data/NewDatabase.kdbx`
- `tests/data/NonAscii.kdbx`

These are fetched from `keepassxreboot/keepassxc` only when
`KPX_REMOTE_FIXTURES=1` is set.

## Refreshing upstream fixtures

When KeePassXC changes upstream and you want to update these fixtures:

1. Pick the upstream commit you want to track.
2. Update the raw GitHub URLs in [`manifest.json`](./manifest.json) to that commit.
3. Recompute each fixture `sha256`.
4. Re-run `KPX_REMOTE_FIXTURES=1 go test ./internal/testcompat -run TestCompatibilityFixtures`.

This gives us "fetch as needed" behavior without making the automated suite
depend on an unstable floating `latest` artifact.

## Recommended next fixtures

- KeePassXC AES-256 + Argon2id
- KeePassXC AES-256 + AES-KDF
- KeePassXC ChaCha20 + Argon2id
- fixtures with custom fields
