package config

import (
	"path/filepath"
	"testing"
)

func TestLoadNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected empty config, got nil")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &Config{
		IgnorePatterns: []string{"*.log", "tmp/"},
		Sync: SyncConfig{
			DefaultTargets: []string{"obsidian"},
			Obsidian: &ObsidianConfig{
				Enabled: true,
				Path:    "/tmp/obsidian",
			},
		},
	}

	if err := cfg.SaveTo(configPath); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(configPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if len(loaded.IgnorePatterns) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(loaded.IgnorePatterns))
	}
	if loaded.Sync.Obsidian == nil {
		t.Fatal("Expected Obsidian config, got nil")
	}
	if !loaded.Sync.Obsidian.Enabled {
		t.Error("Expected Obsidian enabled")
	}
}
