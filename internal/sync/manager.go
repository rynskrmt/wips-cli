package sync

import (
	"context"

	"github.com/rynskrmt/wips-cli/internal/model"
)

// Target is the interface that all sync targets must implement.
type Target interface {
	Name() string
	Sync(ctx context.Context, events []model.WipsEvent) error
}

// Manager handles the synchronization process across multiple targets.
type Manager struct {
	targets map[string]Target
}

// NewManager creates a new Manager instance.
func NewManager() *Manager {
	return &Manager{
		targets: make(map[string]Target),
	}
}

// RegisterTarget registers a new sync target.
func (m *Manager) RegisterTarget(t Target) {
	m.targets[t.Name()] = t
}

// GetTarget retrieves a registered target by name.
func (m *Manager) GetTarget(name string) (Target, bool) {
	t, ok := m.targets[name]
	return t, ok
}
