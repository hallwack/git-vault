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

func TrackCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "track <pattern> [pattern...]",
		Short: "Add one or more file patterns to be encrypted",
		Long: `Adds file patterns to .gitvault.yaml and mirrors them into 
		.gitattributes so Git applies the git-vault filter to matching files.

		Example:
			git-vault track "secrets/*.env" "credentials.json"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrack(args)
		},
	}
}

func runTrack(patterns []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	for _, p := range patterns {
		added := cfg.AddPattern(p)
		if err := gitpkg.AddPattern(p); err != nil {
			return err
		}

		if added {
			fmt.Printf("Tracking pattern: %s\n", patterns)
		} else {
			fmt.Printf("Pattern already tracked: %s\n", patterns)
		}
	}

	return nil
}
