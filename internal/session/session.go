// Package session provides functions for session management and authentication.
package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const sessionFileName = "git-vault-session"

// path returns the location of the session file inside .git/, so it is local
// to the repository and never accidentally committed.
func path() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-dir").Output()

	if err != nil {
		return "", fmt.Errorf("not inside a git repository: %w", err)
	}

	gitDir := strings.TrimSpace(string(out))
	return filepath.Join(gitDir, sessionFileName), nil
}

// Store caches the derived master key for later commands (clean/smudge filters)
// to use without re-prompting for the password.
func Store(key []byte) error {
	path, err := path()
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, key, 0o600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load retrieves the cached master key. Returns an error if no session is
// active (i.e. the repository is locked)
func Load() ([]byte, error) {
	path, err := path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no active session; run 'git-vault unlock' first")
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	return data, nil
}

// Clear removes the cached master key, locking the repository.
func Clear() error {
	path, err := path()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear session file: %w", err)
	}

	return nil
}

// IsActive reports whether a session is currently cached.
func IsActive() bool {
	path, err := path()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}
