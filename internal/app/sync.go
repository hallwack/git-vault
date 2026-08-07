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

// SyncCmd builds the "sync" command. Unlike "track" (which adds a new pattern)
// to both .gitvault.yaml and .gitattributes at once), "sync" goes the other
// direction: it takes whatever is already listed under `patterns` in
// .gitvault.yaml - including patterns added by hand-editing the file - and
// makes sure .gitattributes matches. Git itself only ever reads .gitattributes,
// so this is what actually makes a manually added pattern take effect.
func SyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Reconcile .gitattributes with the patterns listed in .gitvault.yaml",
		Long: `Reads the "patterns" list from .gitvault.yaml and ensures every one of
them has a matching "filter=git-vault" line in .gitattributes.

Useful after manually editing .gitvault.yaml's patterns list, or to repair
.gitattributes if it was accidentally edited or deleted.
Existing lines are left untouched; nothing is removed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync()
		},
	}
}

func runSync() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Patterns) == 0 {
		fmt.Println("No patterns listed in .gitvault.yaml - nothing to sync.")
		return nil
	}

	for _, p := range cfg.Patterns {
		if err := gitpkg.AddPattern(p); err != nil {
			return err
		}
		fmt.Printf("Synced pattern: %s \n", p)
	}

	return nil
}
