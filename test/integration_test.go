package integration

import (
	"os"
	"os/exec"
	"path/filepath"
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
