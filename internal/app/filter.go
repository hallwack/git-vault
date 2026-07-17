/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"github.com/spf13/cobra"
)

// FilterCmd is the parent command for Git clean/smudge filter integration. 
// Git itself will invoke "git-vault filter clean" and "git-vault filter smudge" 
// — see Stage 6 for wiring this up via
// .gitattributes and git config.
func FilterCmd() *cobra.Command {
	filterCmd := &cobra.Command{
		Use:   "filter",
		Short: "Git clean/smudge filter commands (used internally by Git)",
	}

	filterCmd.AddCommand(FilterCleanCmd())
	filterCmd.AddCommand(FilterSmudgeCmd())

	return filterCmd
}
