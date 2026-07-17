/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"git-vault/internal/config"
	gitpkg "git-vault/internal/git"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var force bool

	var initCmd = &cobra.Command{
		Use:   "init [pattern]",
		Short: "Initialize Git Vault in the current repository",
		Long: `Creates a .gitvault.yaml configuration file in the current 
		repository, registers the git-vault clean/smudge filter in this 
		repository's Git config, and optionally adds a pattern to .gitattributes for
		files that should be encrypted (e.g. "secrets/*.env").
		If the file already exists, it will not be overwritten unless the --force flag is used.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var pattern string
			if len(args) == 1 {
				pattern = args[0]
			}
			return runInit(force, pattern)
		},
	}

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing configuration")

	return initCmd
}

func runInit(force bool, pattern string) error {
	if err := ensureInsideGitRepo(); err != nil {
		return err
	}

	if config.Exists() && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", config.FileName)
	}

	cfg := config.Default()
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Created %s\n", config.FileName)

	if err := gitpkg.ConfigureFilter(); err != nil {
		return err
	}
	fmt.Println("Registered git-vault clean/smudge filter in .git/config")

	if pattern != "" {
		if err := gitpkg.AddPattern(pattern); err != nil {
			return err
		}
		fmt.Printf("Added %q to .gitattributes\n", pattern)
	} else {
		fmt.Println("No pattern given - add file patterns to .gitattributes manually, e.g.:")
		fmt.Println("  secrets/*.env filter=git-vault")
	}

	fmt.Println("Next: run 'git-vault unlock' to set a password")

	return nil
}

func ensureInsideGitRepo() error {
	out, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	if err != nil {
		return fmt.Errorf("not a git repository (or git is not installed)")
	}

	if string(out) == "" {
		return fmt.Errorf("not a git repository")
	}

	toplevel, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err == nil {
		root := string(toplevel)
		root = root[:len(root)-1] // trim trailing newline
		cwd, _ := os.Getwd()
		if cwd != "" && root != "" && filepath.Clean(cwd) != filepath.Clean(root) {
			fmt.Printf("Note: repository root is %s (config will be created here in %s)\n", root, cwd)
		}
	}

	return nil

}
