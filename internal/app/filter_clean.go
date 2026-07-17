/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"io"
	"os"

	"git-vault/internal/crypto"
	"git-vault/internal/session"

	"github.com/spf13/cobra"
)

// FilterCleanCmd builds the "filter clean" command. Git calls this with the 
// file's plaintext on stdin and expects the encrypted payload back on stdout. 
// It runs on `git add` / `git commit`.
func FilterCleanCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Encrypt stdin, writing the result to stdout (Git clean filter)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(os.Stdin, os.Stdout)
		},
	}
}

func runClean(in io.Reader, out io.Writer) error {
	// Load key from session
	key, err := session.Load()
	if err != nil {
		// Git vault is locked - refuse rather than silently letting plaintext
		// through unencrypted. This is a security measure to prevent accidental 
		// leaks of sensitive data.
		return fmt.Errorf("cannot encrypt: %w", err)
	}

	// Read .gitattributes file and parse it to get the list of files to encrypt
	plaintext, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	// Generate Nonce
	nonce, err := crypto.GenerateNonce()
	if err != nil {
		return err
	}

	// Encrypt the file contents using the nonce and the master key
	ciphertext, err := crypto.Encrypt(key, nonce, plaintext)
	if err != nil {
		return err
	}

	// Serialize the nonce and ciphertext into a single payload
	payload, err := crypto.Serialize(nonce, ciphertext)
	if err != nil {
		return err
	}

	// Write the payload to the file
	if _, err := out.Write(payload); err != nil {
		return fmt.Errorf("failed to write stdout: %w", err)
	}

	return nil
}
