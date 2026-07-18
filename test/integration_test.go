package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binPath string

// TestMain builds the git-vault binary once before running any test,
// so every test in this package reuses the same compiled binary
// instead of rebuilding per test case.
func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "git-vault-bin")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpDir)

	binPath = filepath.Join(tmpDir, "git-vault")

	cmd := exec.Command("go", "build", "-o", binPath, "./../cmd/git-vault")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic("failed to build git-vault binary: " + err.Error())
	}

	os.Exit(m.Run())
}

func runGitVault(dir string, env []string, args ...string) (string, error) {
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestEndToEndEncryption walks through every test case listed under
// Stage 8 in mvp.md: initialize, unlock, encrypt (via git add),
// commit, checkout, and verify plaintext is restored.
func TestEndToEndEncryption(t *testing.T) {
	dir := t.TempDir()
	const plaintext = "this is a secret\n"
	const password = "correct horse battery staple"

	// --- Test case: Initialize repository ---
	if out, err := runGit(dir, "init"); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	runGit(dir, "config", "user.email", "test@example.com")
	runGit(dir, "config", "user.name", "Test User")

	if out, err := runGitVault(dir, nil, "init", "secret.txt"); err != nil {
		t.Fatalf("git-vault init failed: %v\n%s", err, out)
	}

	// --- Test case: Unlock repository ---
	env := []string{"GIT_VAULT_PASSWORD=" + password}
	if out, err := runGitVault(dir, env, "unlock"); err != nil {
		t.Fatalf("git-vault unlock failed: %v\n%s", err, out)
	}

	// --- Test case: Encrypt file (via `git add`, triggers clean filter) ---
	secretPath := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte(plaintext), 0o644); err != nil {
		t.Fatalf("failed to write plaintext file: %v", err)
	}
	if out, err := runGit(dir, "add", "secret.txt"); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}

	staged, err := runGit(dir, "show", ":secret.txt")
	if err != nil {
		t.Fatalf("git show failed: %v\n%s", err, staged)
	}
	if staged == plaintext {
		t.Fatal("staged content is plaintext — clean filter did not encrypt the file")
	}

	// --- Test case: Commit ---
	if out, err := runGit(dir, "commit", "-m", "add secret"); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}

	// --- Test case: Checkout ---
	if err := os.Remove(secretPath); err != nil {
		t.Fatalf("failed to remove working tree file: %v", err)
	}
	if out, err := runGit(dir, "checkout", "--", "secret.txt"); err != nil {
		t.Fatalf("git checkout failed: %v\n%s", err, out)
	}

	// --- Test case: Restore plaintext ---
	restored, err := os.ReadFile(secretPath)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(restored) != plaintext {
		t.Fatalf("restored content mismatch:\n  want: %q\n  got:  %q", plaintext, string(restored))
	}
}

// TestLockedRepositoryRefusesFilter verifies that Stage 5's safety
// property holds: without an active session, the clean filter must
// fail rather than silently letting plaintext through.
func TestLockedRepositoryRefusesFilter(t *testing.T) {
	dir := t.TempDir()

	if out, err := runGit(dir, "init"); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	runGit(dir, "config", "user.email", "test@example.com")
	runGit(dir, "config", "user.name", "Test User")

	if out, err := runGitVault(dir, nil, "init", "secret.txt"); err != nil {
		t.Fatalf("git-vault init failed: %v\n%s", err, out)
	}
	// Deliberately NOT calling unlock here.

	secretPath := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("should not be committed\n"), 0o644); err != nil {
		t.Fatalf("failed to write plaintext file: %v", err)
	}

	if out, err := runGit(dir, "add", "secret.txt"); err == nil {
		t.Fatalf("expected `git add` to fail while locked, but it succeeded:\n%s", out)
	}
}

// TestInstallAfterClone verifies the fix for the collaboration gap:
// .git/config is local and is never carried over by `git clone`, so
// a fresh clone must NOT already have the git-vault filter
// registered — and running `git-vault install` must register it
// without touching .gitvault.yaml, the salt, or patterns, allowing
// the clone to unlock and decrypt with the same shared password.
func TestInstallAfterClone(t *testing.T) {
	origin := t.TempDir()
	const plaintext = "clone me please\n"
	const password = "shared-team-password"
	env := []string{"GIT_VAULT_PASSWORD=" + password}

	// --- Set up the "origin" repository, as if by the first teammate ---
	if out, err := runGit(origin, "init"); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	runGit(origin, "config", "user.email", "test@example.com")
	runGit(origin, "config", "user.name", "Test User")

	if out, err := runGitVault(origin, nil, "init", "secret.txt"); err != nil {
		t.Fatalf("git-vault init failed: %v\n%s", err, out)
	}
	if out, err := runGitVault(origin, env, "unlock"); err != nil {
		t.Fatalf("git-vault unlock (origin) failed: %v\n%s", err, out)
	}

	if err := os.WriteFile(filepath.Join(origin, "secret.txt"), []byte(plaintext), 0o644); err != nil {
		t.Fatalf("failed to write plaintext file: %v", err)
	}

	// Everything needed by a collaborator must be committed:
	// .gitvault.yaml, .gitattributes, and .gitvault.salt are all
	// required, not just the encrypted file itself. The session
	// file at .git/git-vault-session is correctly excluded — it's
	// outside the working tree and never tracked.
	if out, err := runGit(origin, "add", ".gitvault.yaml", ".gitattributes", ".gitvault.salt", "secret.txt"); err != nil {
		t.Fatalf("git add failed: %v\n%s", err, out)
	}
	if out, err := runGit(origin, "commit", "-m", "add secret"); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}

	// --- Clone into a fresh directory, simulating a teammate's machine ---
	clone := filepath.Join(t.TempDir(), "clone")
	if out, err := exec.Command("git", "clone", origin, clone).CombinedOutput(); err != nil {
		t.Fatalf("git clone failed: %v\n%s", err, out)
	}
	runGit(clone, "config", "user.email", "test@example.com")
	runGit(clone, "config", "user.name", "Test User")

	gitConfigPath := filepath.Join(clone, ".git", "config")

	// --- Sanity check: prove the bug is real — the clone must NOT
	// already have the filter registered, since .git/config is
	// never carried over by `git clone`. ---
	before, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("failed to read cloned .git/config: %v", err)
	}
	if strings.Contains(string(before), "git-vault") {
		t.Fatal("expected cloned .git/config to NOT contain the git-vault filter before running install")
	}

	// --- The fix under test: `install` registers the filter locally ---
	if out, err := runGitVault(clone, nil, "install"); err != nil {
		t.Fatalf("git-vault install failed: %v\n%s", err, out)
	}

	after, err := os.ReadFile(gitConfigPath)
	if err != nil {
		t.Fatalf("failed to read .git/config after install: %v", err)
	}
	if !strings.Contains(string(after), "git-vault") {
		t.Fatal("expected .git/config to contain the git-vault filter after running install")
	}

	// --- Unlock with the shared password and confirm decryption works ---
	if out, err := runGitVault(clone, env, "unlock"); err != nil {
		t.Fatalf("git-vault unlock (clone) failed: %v\n%s", err, out)
	}

	clonedSecretPath := filepath.Join(clone, "secret.txt")
	if err := os.Remove(clonedSecretPath); err != nil {
		t.Fatalf("failed to remove working tree file: %v", err)
	}
	if out, err := runGit(clone, "checkout", "--", "secret.txt"); err != nil {
		t.Fatalf("git checkout failed: %v\n%s", err, out)
	}

	restored, err := os.ReadFile(clonedSecretPath)
	if err != nil {
		t.Fatalf("failed to read restored file: %v", err)
	}
	if string(restored) != plaintext {
		t.Fatalf("restored content mismatch:\n  want: %q\n  got:  %q", plaintext, string(restored))
	}
}
