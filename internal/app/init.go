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

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var force bool

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Git Vault in the current repository",
		Long:  "Creates a .gitvault.yaml configuration file in the current repository. If the file already exists, it will not be overwritten unless the --force flag is used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(force)
		},
	}

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing configuration")

	return initCmd
}

func runInit(force bool) error {
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

	fmt.Printf("Initialized Git Vault configuration: %s\n", config.FileName)
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
