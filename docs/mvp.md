# Git Vault MVP

The goal of the MVP is to build a minimal yet functional Git encryption tool.

**Status: MVP complete (Stage 1–8 all implemented and verified via integration tests).**

## Scope

The MVP focuses on password-based encryption using Git clean/smudge filters.

Not included in this milestone:

- Multi-user support
- Secret scanning
- Repository audit
- Git hooks
- CI/CD integration
- SSH, Age, or GPG key providers

---

# Stage 1 — Project Foundation

## Goal

Create a functional CLI application.

### Tasks

- [x] Initialize Go module
- [x] Configure Cobra CLI
- [x] Implement root command
- [x] Implement version command

---

# Stage 2 — Configuration

## Goal

Create and load project configuration.

### Tasks

- [x] Create `.gitvault.yaml`
- [x] Load configuration
- [x] Validate configuration
- [x] Save configuration

**Addition beyond original scope:** `.gitvault.yaml` also stores `patterns` (file patterns to encrypt, mirrored into `.gitattributes`) and `marker` (an encrypted known-plaintext used to verify a password is correct on unlock, before it's trusted). `Validate()` checks all fields, including structural validation of `marker`.

---

# Stage 3 — Password

## Goal

Generate a master key from a password.

### Tasks

- [x] Prompt password
- [x] Confirm password
- [x] Generate random salt
- [x] Derive master key using Argon2id

**Addition beyond original scope:** password can also be supplied non-interactively via the `GIT_VAULT_PASSWORD` environment variable, needed for automated testing (Stage 8) and as groundwork for headless CI/CD unlock (see roadmap.md v0.6).

---

# Stage 4 — Crypto

## Goal

Encrypt and decrypt arbitrary data.

### Tasks

- [x] Generate nonce
- [x] Encrypt data
- [x] Decrypt data
- [x] Serialize encrypted payload

Cipher: XChaCha20-Poly1305 (AEAD — authenticates data, rejects tampered/corrupted ciphertext rather than silently producing garbage plaintext).

---

# Stage 5 — Git Filter

## Goal

Integrate Git clean/smudge filter.

### Tasks

- [x] Implement clean filter
- [x] Implement smudge filter
- [x] Read stdin
- [x] Write stdout

Implemented as `git-vault filter clean` / `git-vault filter smudge`. Both refuse to run (non-zero exit, no output) if no session is active — locked repositories cannot silently leak plaintext into Git.

---

# Stage 6 — Git Integration

## Goal

Automatically configure Git.

### Tasks

- [x] Configure Git filter
- [x] Update `.gitattributes`
- [x] Validate repository

**Addition beyond original scope:** `.git/config` is local and is never carried over by `git clone`, so a `git-vault install` command was added — it re-registers the filter for collaborators after cloning, without touching `.gitvault.yaml`, salt, or patterns. Also added `git-vault track <pattern...>` to add patterns after the initial `init`, keeping `.gitvault.yaml` and `.gitattributes` in sync.

---

# Stage 7 — Session

## Goal

Prevent repeated password prompts.

### Tasks

- [x] Unlock session
- [x] Store master key
- [x] Retrieve master key
- [x] Lock session

Session key is cached at `.git/git-vault-session` (outside the working tree, never tracked).

---

# Stage 8 — Integration Test

## Goal

Verify end-to-end workflow.

### Test Cases

- [x] Initialize repository
- [x] Unlock repository
- [x] Encrypt file
- [x] Commit
- [x] Checkout
- [x] Restore plaintext

Implemented in `test/integration_test.go`, plus two additional automated tests beyond the original list:

- `TestLockedRepositoryRefusesFilter` — confirms `git add` fails while locked, rather than committing plaintext.
- `TestInstallAfterClone` — confirms `.git/config` is not carried over by `git clone`, and that `git-vault install` fixes it so a cloned collaborator can unlock and decrypt with the shared password.

---

# MVP Completed

The MVP is considered complete now that Git Vault can transparently encrypt and decrypt tracked files during normal Git operations, including across collaborators via `install`, and the full flow is covered by automated integration tests.

## Known limitations carried into the next milestone

- Password verification exists (via the marker), but there is still no key **rotation** — changing the shared password requires manually re-encrypting all tracked files.
- Error handling currently relies on formatted error strings rather than typed/sentinel errors, making programmatic error handling (e.g. in tests or future tooling) less robust than idiomatic Go practice recommends.

See `roadmap.md` for what comes after v0.1.
