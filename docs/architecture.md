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
    filter/
    git/
    password/
    session/
```

---

# Components

## CLI

Responsible for parsing user commands.

Examples:

- init
- unlock
- lock
- version

---

## Config

Responsible for:

- loading configuration
- validating configuration
- saving configuration

---

## Password

Responsible for:

- prompting password
- deriving master key

---

## Crypto

Responsible for:

- encryption
- decryption
- nonce generation

This package has no knowledge of Git.

---

## Session

Responsible for temporarily storing the master key.

---

## Filter

Responsible for implementing Git clean/smudge filters.

Input:

- stdin

Output:

- stdout

---

## Git

Responsible for interacting with Git.

Examples:

- configure filters
- update .gitattributes
- verify repository

---

# Design Principles

- Single Responsibility Principle
- Interface-first design
- Git-independent crypto layer
- Small reusable packages
- No business logic inside CLI
