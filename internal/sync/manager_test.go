package sync

import (
	"context"
	"testing"

	"github.com/rynskrmt/wips-cli/internal/model"
)

type mockTarget struct {
	name      string
	syncCalls int
}

func (m *mockTarget) Name() string {
	return m.name
}

func (m *mockTarget) Sync(ctx context.Context, events []model.WipsEvent) error {
	m.syncCalls++
	return nil
}

func TestManager(t *testing.T) {
	m := NewManager()

	target1 := &mockTarget{name: "target1"}
	target2 := &mockTarget{name: "target2"}

	m.RegisterTarget(target1)
	m.RegisterTarget(target2)

	t1, ok := m.GetTarget("target1")
	if !ok || t1 != target1 {
		t.Error("Failed to get target1")
	}

	t2, ok := m.GetTarget("target2")
	if !ok || t2 != target2 {
		t.Error("Failed to get target2")
	}

	_, ok = m.GetTarget("non-existent")
	if ok {
		t.Error("Should not find non-existent target")
	}
}
