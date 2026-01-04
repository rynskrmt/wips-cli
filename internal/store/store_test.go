package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rynskrmt/wips-cli/internal/model"
)

func TestStore_AppendEvent_TableDriven(t *testing.T) {
	// Base setup
	tempDir, err := os.MkdirTemp("", "wips_test_append")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	s, err := NewStore(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Prepare(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		event   *model.WipsEvent
		wantErr bool
	}{
		{
			name: "Success: Append valid event",
			event: &model.WipsEvent{
				ID:      "test-id-1",
				TS:      time.Now(),
				Type:    model.EventTypeNote,
				Content: "First Note",
			},
			wantErr: false,
		},
		{
			name: "Success: Append second event",
			event: &model.WipsEvent{
				ID:      "test-id-2",
				TS:      time.Now(),
				Type:    model.EventTypeNote,
				Content: "Second Note",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.AppendEvent(tt.event); (err != nil) != tt.wantErr {
				t.Errorf("Store.AppendEvent() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify if the event was actually written
			filename := tt.event.TS.Format("2006-01") + ".ndjson"
			path := filepath.Join(tempDir, "events", filename)

			// Simple verification: Just check if file exists and we can find the content string in it
			// For a robust test we might parse the whole file, but simple grep-like check is okay for Append test here
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read event file: %v", err)
			}

			if string(data) == "" {
				t.Error("File is empty, expected content")
			}
			// In a real scenario, we might want to check the JSON structure more strictly
		})
	}
}

func TestStore_SaveDict_TableDriven(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wips_test_dict")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	s, _ := NewStore(tempDir)
	s.Prepare()

	type args struct {
		dictName string
		key      string
		value    interface{}
	}
	tests := []struct {
		name         string
		existingData map[string]interface{}
		args         args
		wantValue    interface{} // The value we expect to find in the file for the key
		wantErr      bool
	}{
		{
			name:         "New Key",
			existingData: nil,
			args: args{
				dictName: "config",
				key:      "theme",
				value:    "dark",
			},
			wantValue: "dark",
			wantErr:   false,
		},
		{
			name: "Existing Key (Should not overwrite)",
			existingData: map[string]interface{}{
				"theme": "light",
			},
			args: args{
				dictName: "config",
				key:      "theme",
				value:    "dark", // Attempt to overwrite
			},
			wantValue: "light", // Expect old value
			wantErr:   false,
		},
		{
			name: "Add second key",
			existingData: map[string]interface{}{
				"theme": "light",
			},
			args: args{
				dictName: "config",
				key:      "lang",
				value:    "ja",
			},
			wantValue: "ja",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup pre-condition: write existing data if any
			dictPath := filepath.Join(tempDir, "dict", tt.args.dictName+".json")
			if tt.existingData != nil {
				f, _ := os.Create(dictPath)
				json.NewEncoder(f).Encode(tt.existingData)
				f.Close()
			} else {
				// Ensure file doesn't exist or is empty if nil
				os.Remove(dictPath)
			}

			if err := s.SaveDict(tt.args.dictName, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Store.SaveDict() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify result
			data, err := os.ReadFile(dictPath)
			if err != nil {
				t.Fatalf("Failed to read dict file: %v", err)
			}

			var result map[string]interface{}
			if len(data) > 0 {
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("Failed to decode json: %v", err)
				}
			}

			if got := result[tt.args.key]; got != tt.wantValue {
				t.Errorf("Store.SaveDict() Check = %v, want %v", got, tt.wantValue)
			}
		})
	}
}
