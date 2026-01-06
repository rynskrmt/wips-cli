package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/rynskrmt/wips-cli/internal/usecase"
)

func TestRunNote(t *testing.T) {
	// Setup generic temp store
	tempDir, err := os.MkdirTemp("", "wips_test_cmd")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	s, err := store.NewStore(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Prepare(); err != nil {
		t.Fatal(err)
	}

	// Test
	msg := "Integration Test Note"
	u := usecase.NewNoteUsecase(s)
	err = u.RecordNote(msg, tempDir)
	if err != nil {
		t.Errorf("RecordNote() error = %v", err)
	}

	// Verify
	verifyEventExists(t, tempDir, model.EventTypeNote, msg)
}

func TestRunCapture(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wips_test_capture")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	s, err := store.NewStore(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Prepare(); err != nil {
		t.Fatal(err)
	}

	mockGitShow := func() ([]byte, error) {
		return []byte("commit abc1234\n 1 file changed"), nil
	}

	u := usecase.NewCaptureUsecase(s)
	err = u.CaptureEvent("git-commit", tempDir, mockGitShow)
	if err != nil {
		t.Errorf("CaptureEvent() error = %v", err)
	}

	verifyEventExists(t, tempDir, model.EventTypeGitCommit, "commit abc1234\n 1 file changed")
}

func verifyEventExists(t *testing.T, rootDir string, eventType model.EventType, content string) {
	// Find event file
	month := time.Now().Format("2006-01")
	path := filepath.Join(rootDir, "events", month+".ndjson")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read event file: %v", err)
	}

	var event model.WipsEvent
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if event.Type != eventType {
		t.Errorf("Event Type = %v, want %v", event.Type, eventType)
	}
	if event.Content != content {
		t.Errorf("Event Content = %v, want %v", event.Content, content)
	}
}
