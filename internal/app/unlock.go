/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"os"

	"git-vault/internal/config"
	"git-vault/internal/password"
	"git-vault/internal/session"

	"github.com/spf13/cobra"
)

func NewUnlockCmd() *cobra.Command {
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
		return nil
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
			return nil
		}

		if err := os.WriteFile(cfg.SaltFile, salt, 0o600); err != nil {
			return fmt.Errorf("failed to write salt file: %v", err)
		}
	} else {
		_, err := os.ReadFile(cfg.SaltFile)
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
