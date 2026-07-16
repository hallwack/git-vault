# Git Vault Roadmap

## v0.1 — MVP

### Features

- Password authentication
- Argon2id key derivation
- XChaCha20-Poly1305 encryption
- Git clean filter
- Git smudge filter
- Session management
- Automatic Git configuration

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
