# Git Vault Architecture

## Overview

Git Vault uses Git clean/smudge filters to transparently encrypt files before they are stored in Git objects and decrypt them when checked out.

```
Working Tree
      │
      ▼
Git Clean Filter
      │
      ▼
Encryption Engine
      │
      ▼
Git Object Database
      │
      ▼
Remote Repository
```

Checkout performs the reverse process.

```
Remote Repository
      │
      ▼
Git Object Database
      │
      ▼
Git Smudge Filter
      │
      ▼
Decryption Engine
      │
      ▼
Working Tree
```

---

# Project Structure

```
cmd/
    git-vault/

internal/

    app/
    config/
    crypto/
    git/
    password/
    session/

test/
```

**Note:** the original design sketched a standalone `internal/filter/` package for the clean/smudge filter logic. In practice, `filter`, `filter clean`, and `filter smudge` are implemented as thin CLI commands inside `internal/app/` (`filter.go`, `filter_clean.go`, `filter_smudge.go`). Their job is pure orchestration — read stdin, call `session.Load()` and `crypto.Encrypt`/`Decrypt`/`Serialize`/`Deserialize`, write stdout — with no independent logic of their own, so a separate package wasn't warranted. All actual encryption/decryption logic still lives in `internal/crypto`, which remains Git-independent.

---

# Components

## CLI

Responsible for parsing user commands and orchestrating calls into `internal/*` packages. Contains no cryptographic or business logic of its own.

Commands:

- `init [pattern...]` — first-time setup: creates `.gitvault.yaml`, registers the Git filter, records patterns
- `install` — re-registers the Git filter after a fresh clone (see Git component below)
- `unlock` — derives the master key from a password and caches it in the session
- `lock` — clears the cached master key
- `track <pattern...>` — adds file patterns after the initial `init`
- `filter clean` / `filter smudge` — invoked by Git itself, not meant for interactive use
- `version`

---

## Config

Responsible for:

- loading configuration
- validating configuration
- saving configuration

Stores the KDF/cipher scheme (`argon2id`, `xchacha20poly1305`), the list of file `patterns` that should be encrypted (mirrored into `.gitattributes` by the Git component), and a `marker` — an encrypted known-plaintext used to verify a password is correct before it's cached, rather than only discovering a wrong password when a real file fails to decrypt.

---

## Password

Responsible for:

- prompting password (interactively, or via the `GIT_VAULT_PASSWORD` environment variable for automation/testing/future headless CI use)
- confirming password on first setup
- generating a random salt
- deriving master key via Argon2id

---

## Crypto

Responsible for:

- nonce generation
- encryption / decryption (XChaCha20-Poly1305, AEAD)
- serializing/deserializing the on-disk payload format (`[version byte][nonce][ciphertext+tag]`)

This package has no knowledge of Git, files, or the CLI — it only transforms byte slices. It also has no knowledge of Config or Session; callers are responsible for supplying keys and nonces.

---

## Session

Responsible for temporarily storing the master key at `.git/git-vault-session`, outside the working tree so it is never tracked by Git. Used by the filter commands to avoid re-prompting for a password on every Git operation.

---

## Git

Responsible for interacting with Git itself:

- configuring the clean/smudge filter in the repository's **local** `.git/config` (never committed — this is why `install` exists as a separate step for collaborators after cloning)
- updating `.gitattributes` with file patterns
- verifying the current directory is inside a Git repository

---

# Design Principles

- Single Responsibility Principle
- Interface-first design
- Git-independent crypto layer
- Small reusable packages
- No business logic inside CLI
- **Fail closed, not open** — if the filter can't obtain a key (locked), or a password fails marker verification, or ciphertext fails AEAD authentication, the operation errors out rather than silently falling back to unencrypted or corrupted data. `filter.git-vault.required = true` in Git config reinforces this: Git itself aborts if the filter command fails or is missing.
