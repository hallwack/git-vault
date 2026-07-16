# Git Vault

> Transparent Git encryption powered by modern cryptography.

Git Vault is a Git filter that transparently encrypts selected files before they are stored inside Git objects and automatically decrypts them when checking out the repository.

The goal is to keep your working directory readable while ensuring encrypted content is stored in your Git history and remote repository.

> **Project Status**
>
> 🚧 MVP in development.

---

## Features (MVP)

- Transparent Git clean/smudge filter
- Password-based encryption
- Any encryption
- Argon2id key derivation
- YAML configuration
- Git integration

Planned:

- Multi-user support
- SSH / Age / GPG key providers
- Secret scanning
- Repository audit
- Key rotation
- Git hooks
- CI integration

---

## Installation

```bash
go install github.com/hallwack/git-vault@latest
```

---

## Quick Start

Initialize a repository.

```bash
git-vault init
```

Unlock the repository.

```bash
git-vault unlock
```

Lock the repository.

```bash
git-vault lock
```

---

## Commands

```text
git-vault init

git-vault unlock

git-vault lock

git-vault filter clean

git-vault filter smudge

git-vault version
```

---

## Architecture

```
Working Tree
      │
      ▼
 Git Clean Filter
      │
      ▼
Encryption
      │
      ▼
Git Object Database
      │
      ▼
Remote Repository
```

Checkout performs the reverse process using the smudge filter.

---

## Configuration

```yaml
version: 1

algorithm: xchacha20-poly1305

kdf:
  algorithm: argon2id
  salt: RANDOM_SALT
```

---

## Project Structure

```
cmd/
internal/
    app/
    config/
    crypto/
    filter/
    git/
    password/
    session/
```

---

## License

MIT
