// Package password provides functions for password hashing and verification.
package password

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

const (
	saltLen = 16
	keyLen  = 32
	time    = 1
	memory  = 64 * 1024 // 64 MB
	threads = 4
)

// envPasswordVar lets automation (scripts, CI, integration tests) supply a 
// password without an interactive terminal. This mirrors the "headless mode"
// planned for v0.6 in roadmap.md, introduced early here so Stage 8 can be 
// tested without a real TTY.
//
// NOTE: env vars are visible to other processes on the same host (e.g. via 
// /proc on Linux), so this is meant for testing/scripting convenience, not as
// a secure secret-passing mechanism.
const envPasswordVar = "GIT_VAULT_PASSWORD"

// Prompt reads a password from stdin without echoing it to the terminal. If 
// GIT_VAULT_PASSWORD is set, it is used instead and no terminal interaction
// happens at all.
func Prompt(label string) ([]byte, error) {
	if password := os.Getenv(envPasswordVar); password != "" {
		return []byte(password), nil
	}

	fmt.Fprint(os.Stderr, label)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)

	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("password cannot be empty")
	}

	return password, nil
}

// PromptWithConfirmation asks for a password twice and ensure the match.
// Used the first time a repository is unlocked (no salt file yet).
// When GIT_VAULT_PASSWORD is set, both reads resolve to the same value
// automatically, so confirmation trivially passes.
func PromptWithConfirmation() ([]byte, error) {
	password, err := Prompt("Enter Password: ")
	if err != nil {
		return nil, err
	}

	confirmPassword, err := Prompt("Confirm Password: ")
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(password, confirmPassword) {
		return nil, fmt.Errorf("passwords do not match")
	}

	return password, nil
}

// GenerateSalt returns a new random salt for key derivation
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// DeriveKey derives a 32-byte master key from a password and salt
// using Argon2id (see roadmap.md v0.1 - Argon2id key derivation)
func DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, time, memory, threads, keyLen)
}
