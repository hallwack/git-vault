/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"encoding/base64"
	"fmt"
	"os"

	"git-vault/internal/config"
	"git-vault/internal/crypto"
	"git-vault/internal/password"
	"git-vault/internal/session"

	"github.com/spf13/cobra"
)

// markerPlaintext is a fixed known value encrypted on first unlock and
// re-decrypted on every unlock after that. If decryption fails or the recovered
// text does not match, the entered password is wrong - caught immediately,
// instead of only failing later when a real file fails to decrypt.
const markerPlaintext = "git-vault-password-check"

func UnlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unlock",
		Short: "Unlock the repository by deriving and caching the master key",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlock()
		},
	}
}

func runUnlock() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// First unlock ever: no salt file yet, so generate lock and ask the user to
	// confirm their new password. Every unlock after that reuses the existing
	// salt and only asks once.
	firstTime := false
	var salt []byte

	if _, err := os.Stat(cfg.SaltFile); os.IsNotExist(err) {
		firstTime = true
		salt, err = password.GenerateSalt()
		if err != nil {
			return err
		}

		if err := os.WriteFile(cfg.SaltFile, salt, 0o600); err != nil {
			return fmt.Errorf("failed to write salt file: %v", err)
		}
	} else {
		salt, err = os.ReadFile(cfg.SaltFile)
		if err != nil {
			return fmt.Errorf("failed to read salt file: %v", err)
		}
	}

	var pw []byte
	if firstTime {
		fmt.Println("Setting up Git Vault for the first time in this repository.")
		pw, err = password.PromptWithConfirmation()
	} else {
		pw, err = password.Prompt("Enter password: ")
	}
	if err != nil {
		return err
	}

	key := password.DeriveKey(pw, salt)

	if firstTime {
		// No marker yet - create one now so future unlocks can verify the password
		// before trusting it.
		if err := createMarker(&cfg, key); err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return err
		}
	} else {
		if err := verifyMarker(cfg, key); err != nil {
			return err
		}
	}

	// NOTE: there is no way yet to verify the password is *correct* on repeat
	// unlocks (that requires Stage 4 crypto - e.g. trying to decrypt a known
	// marker). For now, a wrong password will simply produce a wrong key and fail
	// later a decrypt time.
	if err := session.Store(key); err != nil {
		return err
	}

	fmt.Println("Git Vault unlocked.")
	return nil
}

// createMarker encrypts markerPlaintext with the freshly derived key and stores
// the result (base64-encoded) into cfg.Marker.
func createMarker(cfg *config.Config, key []byte) error {
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return err
	}

	ciphertext, err := crypto.Encrypt(
		key,
		nonce,
		[]byte(markerPlaintext),
	)
	if err != nil {
		return err
	}

	payload, err := crypto.Serialize(nonce, ciphertext)
	if err != nil {
		return err
	}

	cfg.Marker = base64.StdEncoding.EncodeToString(payload)
	return nil
}

// verifyMarker decrypts cfg.Marker with the given key and confirms it matches
// markerPlaintext. Returns a clear "incorrect password" error if it does not
// - instead of a raw AEAD decryption error.
func verifyMarker(cfg config.Config, key []byte) error {
	if cfg.Marker == "" {
		// Config predates the marker feature (or was hand-edited). Nothing to
		// verify against - proceed without a check rather than blocking unlock
		// entirely.
		return nil
	}

	payload, err := base64.StdEncoding.DecodeString(cfg.Marker)
	if err != nil {
		return fmt.Errorf("failed to decode marker: %w", err)
	}

	nonce, ciphertext, err := crypto.Deserialize(payload)
	if err != nil {
		return fmt.Errorf("failed to read marker: %w", err)
	}

	plaintext, err := crypto.Decrypt(key, nonce, ciphertext)
	if err != nil || string(plaintext) != markerPlaintext {
		return fmt.Errorf("incorrect password (failed to decrypt marker)")
	}

	return nil
}
