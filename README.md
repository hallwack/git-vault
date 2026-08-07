# Git Vault

Git Vault encrypts tracked files transparently using Git's clean/smudge filters. Files stay encrypted in the Git object database — and in your remote, like GitHub — but read normally as plaintext in your local working tree whenever the repository is unlocked.

> **Status:** v0.1 MVP complete. Single shared-password model — see [Limitations](#limitations) before using this for anything beyond a small trusted team.

## Features

- Password-based encryption (Argon2id key derivation)
- XChaCha20-Poly1305 (AEAD) — tampered or corrupted ciphertext is rejected, not silently misdecrypted
- Transparent to normal Git workflow: `git add`, `git commit`, `git checkout` just work
- Password verified on unlock via an encrypted marker, so a typo is caught immediately instead of failing later on a real file
- Session caching, so you don't re-enter your password on every Git operation
- Non-interactive password input (`GIT_VAULT_PASSWORD`) for scripting and CI

## Installation

Build from source (requires Go 1.22+):

```bash
git clone https://github.com/<username>/git-vault.git
cd git-vault
go build -o git-vault ./cmd/git-vault
```

Move the binary somewhere on your `PATH`, or run it directly from the build output.

## Quick start

```bash
# In an existing Git repository:
git-vault init "secrets/*.env"
git-vault unlock          # set a password (asked twice, first time only)

echo "API_KEY=secret" > secrets/.env
git add secrets/.env
git commit -m "add secrets"
git push                  # what lands on the remote is encrypted
```

Teammates cloning the repository:

```bash
git clone <repo-url>
cd <repo>
git-vault install         # registers the filter locally — .git/config is never cloned
git-vault unlock           # same shared password
```

For the full walkthrough — locking, adding more patterns, troubleshooting, CI usage — see [`docs/USAGE.md`](docs/USAGE.md).

## Documentation

| Document                             | Contents                                                                |
| ------------------------------------ | ----------------------------------------------------------------------- |
| [`docs/USAGE.md`](docs/USAGE.md)     | Full usage guide: setup, daily workflow, collaboration, troubleshooting |
| [`architecture.md`](architecture.md) | Package structure and design principles                                 |
| [`mvp.md`](mvp.md)                   | MVP scope and stage-by-stage completion status                          |
| [`roadmap.md`](roadmap.md)           | Planned features beyond v0.1                                            |

## How it works, briefly

Git Vault registers itself as a Git filter. On `git add`/`commit`, Git pipes the file's plaintext to `git-vault filter clean` on stdin and stores whatever comes back on stdout — the encrypted payload. On `checkout`/`pull`, the reverse happens through `git-vault filter smudge`. Both commands read the cached master key from a local session file (`.git/git-vault-session`, never tracked) and refuse to run if the repository is locked, so a missing key can never result in plaintext silently leaking into a commit. See [`architecture.md`](architecture.md) for the full breakdown.

## Testing

```bash
cd test
go test -v ./...
```

Integration tests cover the full encrypt → commit → checkout cycle, that locked repositories refuse to filter, and that a fresh clone can `install` + `unlock` successfully.

## Limitations

This is a v0.1 MVP. Notably:

- **Single shared password** — every collaborator uses the same password. Revoking one person's access means rotating the password and re-encrypting everything; there's no per-user key yet. Public-key support (Age/SSH) is planned for v0.2 — see [`roadmap.md`](roadmap.md).
- No secret scanning, repository audit, or Git hooks yet (v0.3–v0.4).
- No ready-made CI/CD integration, though the non-interactive unlock path it needs already exists (v0.6).

## License
MIT
