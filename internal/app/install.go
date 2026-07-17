/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"git-vault/internal/config"
	gitpkg "git-vault/internal/git"

	"github.com/spf13/cobra"
)

// InstallCmd builds the "install" command. Unlike "init", this never creates
// or modifies .gitvault.yaml, the salt file, or patterns - it only registers
// the clean/smudge filter into this clone's local .git/config, which is never
// committed and therefor must be re-registered by every collaborator after
// cloning.
func InstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Register the git-vault filter in this clone's local Git config",
		Long: `Run this once after cloning a repository that already uses
			Git Vault. It reads the existing .gitvault.yaml (created by whoever
			ran 'git-vault init') and registers the clean/smudge filter in
			.git/config — which lives outside the working tree and is therefore
			never included when the repository is cloned.

			After running this, use 'git-vault unlock' with the shared password
			to start working with encrypted files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall()
		},
	}
}

func runInstall() error {
	if err := gitpkg.ValidateRepository(); err != nil {
		return err
	}

	if !config.Exists() {
		return fmt.Errorf("no %s found - this doesn't look like a Git Vault repository (run 'git-vault init' instead)",
			config.FileName)
	}

	// Load (not just Exists) so a malformed config is caught here, with a clear
	// error, rather than surfacing later inside the clean/smudge filter.
	if _, err := config.Load(); err != nil {
		return err
	}

	if err := gitpkg.ConfigureFilter(); err != nil {
		return err
	}

	fmt.Println("Registered git-vault clean/smudge filter in .git/config")
	fmt.Println("Next: run 'git-vault unlock' with the password shared by your team.")

	return nil
}
