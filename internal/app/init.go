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

func NewInitCmd() *cobra.Command {
	var force bool

	var initCmd = &cobra.Command{
		Use:   "init [pattern...]",
		Short: "Initialize Git Vault in the current repository",
		Long: `Creates a .gitvault.yaml configuration file in the current 
		repository, registers the git-vault clean/smudge filter in this 
		repository's Git config, and records any given file patterns in both
		.gitvault.yaml and .gitattributes so file matching them are encrypted.

		Examples:
			git-vault init
			git-vault init "secrets/*.env"
			git-vault init "secrets/*.env" "credentials.json"
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(force, args)
		},
	}

	initCmd.Flags().BoolVarP(
		&force,
		"force",
		"f",
		false,
		"overwrite existing configuration",
	)

	return initCmd
}

func runInit(force bool, patterns []string) error {
	if err := gitpkg.ValidateRepository(); err != nil {
		return err
	}

	if config.Exists() && !force {
		return fmt.Errorf(
			"%s already exists (use --force to overwrite)",
			config.FileName,
		)
	}

	cfg := config.Default()
	for _, p := range patterns {
		cfg.AddPattern(p)
	}

	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Created %s\n", config.FileName)

	if err := gitpkg.ConfigureFilter(); err != nil {
		return err
	}
	fmt.Println("Registered git-vault clean/smudge filter in .git/config")

	if len(patterns) == 0 {
		fmt.Println("No file patterns given - add file patterns later with 'git-vault track <pattern>'")
	} else {
		for _, p := range patterns {
			if err := gitpkg.AddPattern(p); err != nil {
				return err
			}
			fmt.Printf("Tracking pattern: %s\n", p)
		}
	}

	fmt.Println("Next: run 'git-vault unlock' to set a password")

	return nil
}
