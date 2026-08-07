/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewRootCmd builds the root command and registers all subcommands.
// No business logic lives here - each subcommand delegates to its
// own file in internal/app/, which in turn will delegate to the
// revelant internal/* package (config, crypto, password, etc)
func NewRootCmd() *cobra.Command {

	var rootCmd = &cobra.Command{
		Use:   "git-vault",
		Short: "Git Vault - transparent file encryption for Git",
		Long: `Git Vault encrypts tracked files transparently using Git
		clean/smudge filters, so secrets stay encrypted in the repository
		and are only decrypted in your working tree.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(InitCmd())
	rootCmd.AddCommand(VersionCmd())
	rootCmd.AddCommand(ValidateCmd())
	rootCmd.AddCommand(UnlockCmd())
	rootCmd.AddCommand(LockCmd())
	rootCmd.AddCommand(FilterCmd())
	rootCmd.AddCommand(InstallCmd())
	rootCmd.AddCommand(TrackCmd())
	rootCmd.AddCommand(SyncCmd())

	return rootCmd
}

// Execute runs the root command and handless top-level errors.
// This is the single entry point called from cmd/git-vault/main.go.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		os.Exit(1)
	}
}
