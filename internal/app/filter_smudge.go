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

// FilterSmudgeCmd builds the "filter smudge" command. Git calls this with the
// stored (encrypted) payload on stdin and expects plaintext back on stdout.
// It runs on `git checkout` / clone / pull.
func FilterSmudgeCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "smudge",
		Short:  "Process files after they are checked out from the repository",
		Hidden: true, // internal command, not meant for interactive use
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSmudge(os.Stdin, os.Stdout)
		},
	}
}

func runSmudge(in io.Reader, out io.Writer) error {
	// Load key from session
	key, err := session.Load()
	if err != nil {
		// Git vault is locked - refuse rather than writing garbage/undecrypted
		// into the working tree.
		return fmt.Errorf("cannot decrypt: %w", err)
	}

	// Read the payload from stdin
	payload, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	// Crypto Deserialize the payload to get the nonce and ciphertext
	nonce, ciphertext, err := crypto.Deserialize(payload)
	if err != nil {
		return err
	}

	// Decrypt the ciphertext using the nonce and the master key
	plaintext, err := crypto.Decrypt(key, nonce, ciphertext)
	if err != nil {
		return err
	}

	// Write the plaintext to stdout
	if _, err := out.Write(plaintext); err != nil {
		return fmt.Errorf("failed to write stdout; %w", err)
	}

	return nil
}
