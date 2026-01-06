package obsidian

import (
	"strings"
	"testing"
	"time"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/rynskrmt/wips-cli/internal/model"
)

func TestFormatFilename(t *testing.T) {
	date := time.Date(2023, 10, 25, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		format   string
		expected string
	}{
		{"{{yyyy}}-{{mm}}-{{dd}}.md", "2023-10-25.md"},
		{"Daily/{{yyyy}}/{{mm}}/{{dd}}", "Daily/2023/10/25"},
		{"Context-{{yyyy}}-{{mm}}.md", "Context-2023-10.md"},
	}

	for _, tt := range tests {
		got := formatFilename(tt.format, date)
		if got != tt.expected {
			t.Errorf("formatFilename(%q) = %q, want %q", tt.format, got, tt.expected)
		}
	}
}

func TestUpdateFileContent(t *testing.T) {
	cfg := &config.ObsidianConfig{
		SectionHeader: "## wips logs",
		AppendAt:      "bottom",
	}
	target := NewTarget(cfg, nil, TargetOptions{})

	newSection := "## wips logs\n\n- log 1\n- log 2"

	tests := []struct {
		name     string
		existing string
		expected string
	}{
		{
			name:     "Empty file",
			existing: "",
			expected: "## wips logs\n\n- log 1\n- log 2\n",
		},
		{
			name:     "Append to bottom",
			existing: "# Daily Note\n\nSome content\n",
			expected: "# Daily Note\n\nSome content\n\n## wips logs\n\n- log 1\n- log 2\n",
		},
		{
			name:     "Replace existing section",
			existing: "# Daily Note\n\n## wips logs\n- old log\n\n## Next Section\n",
			expected: "# Daily Note\n\n## wips logs\n\n- log 1\n- log 2\n\n## Next Section\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := target.updateFileContent(tt.existing, newSection)
			// Normalize newlines for comparison
			got = strings.TrimSpace(got)
			expected := strings.TrimSpace(tt.expected)
			if got != expected {
				t.Errorf("updateFileContent() got:\n%q\nwant:\n%q", got, expected)
			}
		})
	}
}

// mockStore implements store.Store interface for testing
type mockStore struct {
	dicts map[string]map[string]interface{}
}

func (m *mockStore) Prepare() error                                                { return nil }
func (m *mockStore) AppendEvent(event *model.WipsEvent) error                      { return nil }
func (m *mockStore) SaveDict(dictName string, key string, value interface{}) error { return nil }
func (m *mockStore) LoadDict(dictName string) (map[string]interface{}, error) {
	if m.dicts == nil {
		return nil, nil
	}
	return m.dicts[dictName], nil
}
func (m *mockStore) GetEvents(start, end time.Time) ([]model.WipsEvent, error) { return nil, nil }
func (m *mockStore) UpdateEvent(id string, mutator func(*model.WipsEvent) error) error {
	return nil
}
func (m *mockStore) DeleteEvent(id string) error { return nil }
func (m *mockStore) GetRootDir() string          { return "" }

func TestGenerateContent(t *testing.T) {
	mockS := &mockStore{
		dicts: map[string]map[string]interface{}{
			"dirs": {
				"dir1": "/path/to/project",
			},
			"repos": {
				"repo1": map[string]interface{}{"name": "wips-cli"},
			},
		},
	}

	cfg := &config.ObsidianConfig{
		SectionHeader: "## wips logs",
	}
	target := NewTarget(cfg, mockS, TargetOptions{})

	now := time.Now()
	repoID := "repo1"
	cwdID := "dir1"

	events := []model.WipsEvent{
		{
			TS:      now,
			Content: "a1b2c3d feat: add feature",
			Type:    model.EventTypeGitCommit,
			Ctx:     model.Context{RepoID: &repoID},
		},
		{
			TS:      now.Add(1 * time.Minute),
			Content: "working on something",
			Type:    model.EventTypeNote,
			Ctx:     model.Context{CwdID: &cwdID},
		},
	}

	content, err := target.generateContent(now, events)
	if err != nil {
		t.Fatalf("generateContent failed: %v", err)
	}

	// Verify content
	if !strings.Contains(content, "## wips logs") {
		t.Error("Content missing section header")
	}
	if !strings.Contains(content, "### @wips-cli") {
		t.Error("Content missing repo header")
	}
	if !strings.Contains(content, "### 📁 /path/to/project") {
		t.Error("Content missing dir header")
	}
	if !strings.Contains(content, "feat: add feature") {
		t.Error("Content missing commit message")
	}
	if !strings.Contains(content, "working on something") {
		t.Error("Content missing note message")
	}
}
