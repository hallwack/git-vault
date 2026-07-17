package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const attributesFileName = ".gitattributes"

func ValidateRepository() error {
	out, err := exec.Command(
		"git",
		"rev-parse",
		"--is-inside-work-tree",
	).Output()

	if err != nil {
		return fmt.Errorf("not a git repository (or git is not installed)")
	}

	if strings.TrimSpace(string(out)) != "true" {
		return fmt.Errorf("not a git repository")
	}

	return nil
}

// ConfigureFilter registers git-vault as a Git clean/smudge filter named
// "git-vault", writing the config into the repository's local .git/config
// (not the global git config)
//
// "required = true" makes Git abort the operation if the filter command fails
// or is missing, instead of silently passing the file through unfiltered
// - critical so a broken filter never results in plaintext being commited by
// accident.
func ConfigureFilter() error {
	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve git-vault binary path: %w", err)
	}

	binPath, err = filepath.Abs(binPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute binary path: %w", err)
	}

	settings := map[string]string{
		"filter.git-vault.clean":    fmt.Sprintf("%q filter clean", binPath),
		"filter.git-vault.smudge":   fmt.Sprintf("%q filter smudge", binPath),
		"filter.git-vault.required": "true",
	}

	for key, value := range settings {
		if err := runGitConfig(key, value); err != nil {
			return err
		}
	}

	return nil
}

func runGitConfig(key, value string) error {
	cmd := exec.Command("git", "config", "--local", key, value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set git config %s: %w (%s)", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// AddPattern appends a filter rule to .gitattributes for the given file pattern, e.g. AddPattern("secret/*.env") writes:
//
// secrets/*.env filter=git-vault
//
// It is idempotent: calling it again with the same pattern is a no-op rather
// that duplicating the line.
func AddPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	line := fmt.Sprintf("%s filter=git-vault", pattern)

	existing, err := os.ReadFile(attributesFileName)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", attributesFileName, err)
	}

	for _, l := range strings.Split(string(existing), "\n") {
		if strings.TrimSpace(l) == line {
			return nil // already present, nothing to do
		}
	}

	f, err := os.OpenFile(attributesFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", attributesFileName, err)
	}
	defer f.Close()

	if _, err := f.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write to %s: %w", attributesFileName, err)
	}

	return nil
}
