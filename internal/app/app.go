// Package app provides application-wide initialization and dependency management.
package app

import (
	"fmt"
	"os"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/rynskrmt/wips-cli/internal/store"
)

// App holds the application-wide dependencies.
// It provides a centralized way to access common resources
// like the data store and configuration.
type App struct {
	Store  store.Store
	Config *config.Config
}

// New creates a new App instance with initialized dependencies.
// It performs the following steps:
// 1. Loads the configuration.
// 2. Initializes the data store (using WIPS_HOME env var if set, otherwise defaults).
// 3. Prepares the store (creates necessary directories).
//
// Returns an error if any initialization step fails.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	s, err := store.NewStore(os.Getenv("WIPS_HOME"))
	if err != nil {
		return nil, fmt.Errorf("failed to init store: %w", err)
	}

	if err := s.Prepare(); err != nil {
		return nil, fmt.Errorf("failed to prepare store: %w", err)
	}

	return &App{
		Store:  s,
		Config: cfg,
	}, nil
}

// NewWithStore creates a new App instance with a custom store.
// This is useful for testing with mock stores.
func NewWithStore(s store.Store) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &App{
		Store:  s,
		Config: cfg,
	}, nil
}

// HiddenDirs returns the list of hidden directories from config.
// Returns an empty slice if no hidden directories are configured.
func (a *App) HiddenDirs() []string {
	if a.Config == nil {
		return nil
	}
	return a.Config.HiddenDirectories
}
