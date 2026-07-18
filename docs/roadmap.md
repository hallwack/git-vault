# Git Vault Roadmap

## v0.1 — MVP ✅ Complete

### Features

- [x] Password authentication
- [x] Argon2id key derivation
- [x] XChaCha20-Poly1305 encryption
- [x] Git clean filter
- [x] Git smudge filter
- [x] Session management
- [x] Automatic Git configuration

Also delivered, beyond the original v0.1 feature list (see `mvp.md` for details):

- Password verification via an encrypted marker (fail fast on wrong password, instead of only failing at decrypt time)
- `git-vault install` — fixes the collaboration gap where `.git/config` is never carried over by `git clone`
- `git-vault track` — add file patterns after initial setup
- Non-interactive password input via `GIT_VAULT_PASSWORD` (early groundwork for v0.6's headless mode)
- Automated integration test suite covering the full encrypt/commit/checkout cycle, locked-repository safety, and the clone/install flow

---

## v0.2 — Multi-user

### Features

- Public key support
- Multiple recipients
- Age provider
- SSH provider
- Recipient management

Commands:

```
git-vault key add
git-vault key remove
```

**Note:** v0.1 ships with a single shared password model. If a team member leaves, the only way to revoke their access is to rotate the password and re-encrypt all tracked files — there is no way to remove just one person's access without doing so. This is the main motivation for prioritizing v0.2's asymmetric (public key) model.

---

## v0.3 — Repository Security

### Features

- Secret scanning
- Repository audit
- Doctor command
- Repository health report

Commands:

```
git-vault scan

git-vault audit

git-vault doctor
```

**Note:** `doctor`/`audit` are a natural fit for checking that `.gitvault.yaml`'s `patterns` list, `.gitattributes`, and `.git/config` are all consistent with each other — a class of bug the v0.1 `install` fix was addressing manually.

---

## v0.4 — Git Hooks

### Features

- Pre-commit hook
- Post-checkout hook
- Automatic scanning
- Automatic encryption suggestions

---

## v0.5 — Key Management

### Features

- Key rotation
- Algorithm migration
- Compression
- Metadata verification

---

## v0.6 — CI/CD

### Features

- GitHub Actions
- GitLab CI
- Unlock using secrets
- Headless mode

**Note:** the non-interactive `GIT_VAULT_PASSWORD` path added in v0.1 already covers the core requirement (`unlock` without a TTY). This milestone is mainly about platform integration — documented recipes for GitHub Actions Secrets / GitLab CI Variables, and possibly a `--password-file` option as a more secure alternative to a bare env var.

---

## v1.0

### Stable Release

Goals:

- Stable CLI
- Stable configuration format
- Cross-platform support
- Full documentation
- Unit tests
- Integration tests
- Performance benchmarks

---

# Future Ideas

Potential features after v1.0:

- Windows Credential Manager
- macOS Keychain
- Linux Secret Service
- Hardware Security Keys
- YubiKey
- HashiCorp Vault
- AWS KMS
- Azure Key Vault
- Google Cloud KMS
- Plugin system
