# Git Vault MVP

The goal of the MVP is to build a minimal yet functional Git encryption tool.

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

---

# Stage 3 — Password

## Goal

Generate a master key from a password.

### Tasks

- [ ] Prompt password
- [ ] Confirm password
- [ ] Generate random salt
- [ ] Derive master key using Argon2id

---

# Stage 4 — Crypto

## Goal

Encrypt and decrypt arbitrary data.

### Tasks

- [ ] Generate nonce
- [ ] Encrypt data
- [ ] Decrypt data
- [ ] Serialize encrypted payload

---

# Stage 5 — Git Filter

## Goal

Integrate Git clean/smudge filter.

### Tasks

- [ ] Implement clean filter
- [ ] Implement smudge filter
- [ ] Read stdin
- [ ] Write stdout

---

# Stage 6 — Git Integration

## Goal

Automatically configure Git.

### Tasks

- [ ] Configure Git filter
- [ ] Update `.gitattributes`
- [ ] Validate repository

---

# Stage 7 — Session

## Goal

Prevent repeated password prompts.

### Tasks

- [ ] Unlock session
- [ ] Store master key
- [ ] Retrieve master key
- [ ] Lock session

---

# Stage 8 — Integration Test

## Goal

Verify end-to-end workflow.

### Test Cases

- [ ] Initialize repository
- [ ] Unlock repository
- [ ] Encrypt file
- [ ] Commit
- [ ] Checkout
- [ ] Restore plaintext

---

# MVP Completed

The MVP is considered complete when Git Vault can transparently encrypt and decrypt tracked files during normal Git operations.
