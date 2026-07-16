/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const configFileName = ".gitvault.yaml"

// Config is the minimal shape of .gitvault.yaml for Stage 1/2.
// This will move to internal/config once that package is built out
// (see mvp.md Stage 2); init.go should then call config.Save(cfg)
// instead of marshaling YAML directly.
type Config struct {
	Version  int    `yaml:"version"`
	KDF      string `yaml:"kdf"`
	Cipher   string `yaml:"cipher"`
	SaltFile string `yaml:"salt_file"`
}

func defaultConfig() Config {
	return Config{
		Version:  1,
		KDF:      "argon2id",
		Cipher:   "aes-256-gcm",
		SaltFile: ".gitvault.salt",
	}
}

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

	if _, err := os.Stat(configFileName); err == nil && !force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", configFileName)
	}

	cfg := defaultConfig()

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(configFileName, data, 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", configFileName, err)
	}

	fmt.Printf("Initialized Git Vault configuration: %s\n", configFileName)
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
