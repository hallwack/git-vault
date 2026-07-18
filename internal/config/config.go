// Package config provides the configuration for the application.
package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const FileName = ".gitvault.yaml"

// minMarkerPayloadLen mirrors crypto.NonceSize (24) + 1 version byte.
// Duplicated here (rather than importing internal/crypto) so this package can
// validate structural shape without depending on the crypto package's internals
// - config only need to know "is this plausibly a serialized payload", not how
// to decrypt it.
const minMarkerPayloadLen = 1 + 24

// Config is the parsed shape of .gitvault.yaml.
type Config struct {
	Version  int      `yaml:"version"`
	KDF      string   `yaml:"kdf"`
	Cipher   string   `yaml:"cipher"`
	SaltFile string   `yaml:"salt_file"`
	Patterns []string `yaml:"patterns"`

	// Marker is a base64-encoded, encrypted payload of a known plaintext. Set
	// once on first unlock, then used on every subsequent unlock to verify the
	// entered password is correct before caching it - instead of only failing
	// later at decrypt time on a real file.
	Marker string `yaml:"marker,omitempty"`
}

// Supported algorithm values, checked by Validate.
var (
	supportedKDFs = map[string]bool{
		"argon2id": true,
		"bcrypt":   true,
	}
	supportedCiphers = map[string]bool{
		"xchaxchacha20poly1305": true,
		"aes256gcm":             true,
		"aes256ctr":             true,
	}
)

// Default returns a Config populated with the values used at `init` time.
func Default() Config {
	return Config{
		Version:  1,
		KDF:      "argon2id",
		Cipher:   "xchaxchacha20poly1305",
		SaltFile: ".gitvault.salt",
		Patterns: []string{},
	}
}

// Exists reports whether a config file is already present in the current
// directory.
func Exists() bool {
	_, err := os.Stat(FileName)
	return err == nil
}

// Load reads and parses .gitvault.yaml from the current directory, then
// validates its contents.
func Load() (Config, error) {
	var cfg Config

	data, err := os.ReadFile(FileName)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, fmt.Errorf("no %s found; run 'git-vault init' first", FileName)
		}
		return cfg, fmt.Errorf("failed to parse %s: %w", FileName, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse %s: %w", FileName, err)
	}

	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid %s: %w", FileName, err)
	}

	return cfg, nil
}

// Save validates and writes the config to .gitvault.yaml in the current
// directory
func Save(cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(FileName, data, 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", FileName, err)
	}

	return nil
}

// AddPattern appends a file pattern to the config if not already present.
// Returns true if the pattern was newly added, false if it was already there.
// Caller is responsible for persisting via Save and for mirroring the pattern
// into .gitattributes (see internal/git.AddPattern) - this only updates in the
// in-memory struct.
func (c *Config) AddPattern(pattern string) bool {
	if slices.Contains(c.Patterns, pattern) {
		return false
	}

	c.Patterns = append(c.Patterns, pattern)
	return true
}

// Validate checks that all required fields are present and hold values
// Git Vault actually supports. Called automatically by both Load and Save, so
// a malformed config is caught the moment it is read or written, not later when
// some command tries to use it.
func (c Config) Validate() error {
	if c.Version <= 0 {
		return fmt.Errorf("version must be a positive integer")
	}

	if c.KDF == "" {
		return fmt.Errorf("kdf is required")
	}

	if !supportedKDFs[c.KDF] {
		var kdfs []string
		for kdf := range supportedKDFs {
			kdfs = append(kdfs, kdf)
		}
		supportedStr := strings.Join(kdfs, ", ")
		return fmt.Errorf("unsupported kdf %q (supported: %v)", c.KDF, supportedStr)
	}

	if c.Cipher == "" {
		return fmt.Errorf("cipher is required")
	}

	if !supportedCiphers[c.Cipher] {
		var chipers []string
		for chiper := range supportedCiphers {
			chipers = append(chipers, chiper)
		}
		supportedStr := strings.Join(chipers, ", ")
		return fmt.Errorf("unsupported cipher %q (supported: %v)", c.Cipher, supportedStr)
	}

	if c.SaltFile == "" {
		return fmt.Errorf("salt_file is required")
	}

	for i, p := range c.Patterns {
		if p == "" {
			return fmt.Errorf("patterns[%d] is empty", i)
		}
	}

	if c.Marker != "" {
		raw, err := base64.StdEncoding.DecodeString(c.Marker)
		if err != nil {
			return fmt.Errorf("marker is not valid base64: %w", err)
		}

		if len(raw) < minMarkerPayloadLen {
			return fmt.Errorf("marker payload too short: got %d bytes, want at least %d",
				len(raw), minMarkerPayloadLen)
		}
	}

	return nil
}
