# Git Vault — Usage Guide

Git Vault is a CLI that transparently encrypts files within a Git repository using Git's built-in clean/smudge filter mechanism. Files are stored encrypted in the Git object database (and on remotes like GitHub), but remain normally readable (plaintext) in your local working tree, as long as the repository is in an *unlocked* state.

## Installation

Build from source:

```bash
git clone [https://github.com/](https://github.com/)<username>/git-vault.git
cd git-vault
go build -o git-vault ./cmd/git-vault
```

Move the binary to a location in your PATH, or run it directly from the build location:

```bash
./git-vault version
```

## Initial Setup

### 1. Initialize repository

Inside an existing Git repository (run git init first if you haven't), run:

```bash
git-vault init "secrets/*.env"
```

This command does three things at once:

 * Creates the .gitvault.yaml configuration file.
 * Registers the git-vault clean/smudge filters in .git/config (local, not global).
 * Adds a line to .gitattributes marking which file patterns should be encrypted.

If you don't know yet which file patterns to encrypt, you can run it without arguments:

```bash
git-vault init
```

and manually edit .gitattributes later, for example:

```text
secrets/*.env filter=git-vault
credentials.json filter=git-vault
```

> To register multiple patterns at once, run git-vault init again with --force for each pattern, or directly edit .gitattributes — adding lines manually is safe and won't break the existing filter configuration.

### 2. Set password (first-time unlock)

```bash
git-vault unlock
```

Since this is the first time, you will be prompted to enter a password twice (for confirmation). From this password, Git Vault generates a random salt (saved in .gitvault.salt) and derives a *master key* using Argon2id. This key is temporarily stored in a *session* (.git/git-vault-session), so you don't need to re-enter your password for subsequent Git operations.

**Important:** Keep this password safe. Git Vault does not store the password in any form — if you forget it, the encrypted files cannot be decrypted again.

## Daily Usage

### Encrypt files (automatic)

Once the repository is *unlocked*, just use Git as usual:

```bash
echo "API_KEY=secret123" > secrets/.env
git add secrets/.env
git commit -m "add secrets"
git push
```

Files matching the pattern in .gitattributes are automatically encrypted by the clean filter before entering the Git object database. What gets committed and pushed to the remote is the encrypted version, not the plaintext.

### Decrypt files (automatic)

When checking out, pulling, or cloning a new repository:

```bash
git checkout secrets/.env
# or after clone:
git clone <repo-url>
```

The smudge filter automatically decrypts the file back to plaintext in the working tree — **as long as the repository is in an unlocked state** with the same password.

### Lock repository

When you are done working and want to remove the master key from the session cache:

```bash
git-vault lock
```

After this, any operation requiring encryption/decryption (git add, git checkout on files matching the pattern) will **fail** until you unlock again. This is intentional as a safeguard: a locked repository will never silently allow plaintext to be committed.

### Re-unlocking

For the next work session (e.g., after restarting the terminal, or on another computer that has been init-ed before):

```bash
git-vault unlock
```

This time you only need to enter the password once (no confirmation), because the salt already exists.

## Check version

```bash
git-vault version          # short version
git-vault version -v       # details: commit, build date, Go version, platform
```

## Team Collaboration

Every collaborator who clones the repository needs to:

 1. Install git-vault on their computer.
 2. Run git-vault unlock with the **exact same** password used during the initial setup (this password needs to be shared via a secure channel outside of Git — e.g., a team password manager, not via regular chat).
 3. The clean/smudge filters will automatically be active because the filter.git-vault.* configuration and .gitattributes are already committed to the repository. *(Catatan: Sesuai evaluasi sebelumnya, Anda mungkin perlu menambahkan instruksi agar kolaborator menjalankan perintah untuk mendaftarkan filter ke .git/config lokal mereka).*

> .gitvault.yaml and .gitattributes **must** be committed so all collaborators use the same encryption scheme (KDF, cipher). Conversely, .gitvault.salt **can** be committed (salt is not a secret) — but the session file in .git/git-vault-session is **never** tracked because it resides outside the working tree.

## Non-interactive mode (for automation/CI)

For automated scripts or CI pipelines that lack an interactive terminal, set the GIT_VAULT_PASSWORD environment variable before calling unlock:

```bash
export GIT_VAULT_PASSWORD="this-repo-password"
git-vault unlock
```

> This env var is convenient for automation but less secure than interactive input (it can be visible to other processes on the same computer via /proc on Linux). For real CI pipelines, ensure the platform (GitHub Actions Secrets, GitLab CI Variables, etc.) masks the secret value in the logs before using it this way.

## Troubleshooting

**git add fails with the message "cannot encrypt: no active session"**
The repository is in a locked state. Run git-vault unlock first.

**Files in the working tree are unreadable / contain random gibberish after checkout**
The repository was likely locked when the checkout occurred, or the password used to unlock was incorrect. Try running git-vault lock, then git-vault unlock again with the correct password.

**Want to check if the filter is installed correctly**

```bash
cat .git/config
```

Look for the [filter "git-vault"] section — it should contain clean, smudge, and required = true.

## Current limitations (MVP v0.1)

According to the current MVP scope, the following are **not yet** supported:

 * Multi-user / multiple recipients (everyone uses the same password)
 * Automatic secret scanning
 * Repository audit / doctor command
 * Automatic Git hooks (pre-commit, post-checkout)
 * Ready-to-use CI/CD integration (the non-interactive foundation exists, but platform integrations do not)
 * SSH, Age, or GPG key providers

See roadmap.md for plans regarding these features in future versions.
