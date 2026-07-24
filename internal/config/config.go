// Package config loads the kvex YAML configuration file describing the
// set of Key Vaults the user wants to browse.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Vault is a single named Key Vault entry from the config file.
type Vault struct {
	Name         string `yaml:"name"`
	URI          string `yaml:"uri"`
	Subscription string `yaml:"subscription,omitempty"`
}

// Config is the top-level kvex configuration.
type Config struct {
	Vaults []Vault `yaml:"vaults"`
}

// DefaultPath returns ~/.config/kvex/config.yaml.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "kvex", "config.yaml"), nil
}

// Load reads and parses the config file at path, validating that every
// vault has a name and URI.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if len(cfg.Vaults) == 0 {
		return nil, fmt.Errorf("config %s defines no vaults", path)
	}

	seen := make(map[string]struct{}, len(cfg.Vaults))
	for i, v := range cfg.Vaults {
		if v.Name == "" {
			return nil, fmt.Errorf("vault at index %d has no name", i)
		}
		if v.URI == "" {
			return nil, fmt.Errorf("vault %q has no uri", v.Name)
		}
		if _, dup := seen[v.Name]; dup {
			return nil, fmt.Errorf("duplicate vault name %q", v.Name)
		}
		seen[v.Name] = struct{}{}
	}

	return &cfg, nil
}
