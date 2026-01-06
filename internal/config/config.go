package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	IgnorePatterns    []string   `toml:"ignore_patterns"`
	HiddenDirectories []string   `toml:"hidden_directories"`
	Sync              SyncConfig `toml:"sync"`
}

type SyncConfig struct {
	DefaultTargets []string        `toml:"default_targets"`
	Obsidian       *ObsidianConfig `toml:"obsidian,omitempty"`
}

type ObsidianConfig struct {
	Enabled             bool   `toml:"enabled"`
	Path                string `toml:"path"` // Absolute path to Daily Notes folder
	DailyFilenameFormat string `toml:"daily_filename_format"`
	SectionHeader       string `toml:"section_header"`
	AppendAt            string `toml:"append_at"` // "top" or "bottom"
	SummaryFormat       string `toml:"summary_format"`
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wip", "config.toml"), nil
}

func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(configPath)
	if os.IsNotExist(err) {
		return &Config{}, nil // Return empty config if not exists
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := toml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to file.
func (c *Config) Save() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}

// AddHiddenDir adds a directory to the hidden list.
// The path is normalized to an absolute path.
func (c *Config) AddHiddenDir(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if already exists
	for _, d := range c.HiddenDirectories {
		if d == absPath {
			return nil // Already exists
		}
	}

	c.HiddenDirectories = append(c.HiddenDirectories, absPath)
	return c.Save()
}

// RemoveHiddenDir removes a directory from the hidden list.
func (c *Config) RemoveHiddenDir(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	newList := make([]string, 0, len(c.HiddenDirectories))
	found := false
	for _, d := range c.HiddenDirectories {
		if d == absPath {
			found = true
			continue
		}
		newList = append(newList, d)
	}

	if !found {
		return fmt.Errorf("directory not found in hidden list: %s", absPath)
	}

	c.HiddenDirectories = newList
	return c.Save()
}

// IsHiddenDir checks if a path is under a hidden directory.
func (c *Config) IsHiddenDir(path string) bool {
	for _, hiddenDir := range c.HiddenDirectories {
		// Check if path starts with hiddenDir (is under the hidden directory)
		if strings.HasPrefix(path, hiddenDir) {
			// Ensure it's actually a subdirectory, not just a prefix match
			if path == hiddenDir || strings.HasPrefix(path, hiddenDir+string(filepath.Separator)) {
				return true
			}
		}
	}
	return false
}
