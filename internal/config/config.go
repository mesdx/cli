package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "config.json"

// DefaultMemoryDir is the default directory for storing memory markdown files.
const DefaultMemoryDir = "docs/mesdx-memory"

// Config represents the mesdx configuration.
type Config struct {
	RepoRoot    string   `json:"repoRoot"`
	SourceRoots []string `json:"sourceRoots"`
	MemoryDir   string   `json:"memoryDir,omitempty"`
}

// ConfigPath returns the path to the config file in the mesdx directory.
func ConfigPath(mesdxDir string) string {
	return filepath.Join(mesdxDir, configFileName)
}

// Save writes the configuration to disk.
func Save(cfg *Config, mesdxDir string) error {
	configPath := ConfigPath(mesdxDir)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Load reads the configuration from disk.
func Load(mesdxDir string) (*Config, error) {
	configPath := ConfigPath(mesdxDir)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
