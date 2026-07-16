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

// Prompt reads a password from the terminal without echoing it back to the screen.
func Prompt(label string) ([]byte, error) {
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
