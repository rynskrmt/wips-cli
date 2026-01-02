package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWipsEvent_JSON(t *testing.T) {
	// Fixed time for stable testing
	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		event    WipsEvent
		wantJSON string
	}{
		{
			name: "Basic Note Event",
			event: WipsEvent{
				ID:      "test-id",
				TS:      fixedTime,
				Type:    EventTypeNote,
				Content: "Hello World",
			},
			// time.Time marshals to RFC3339 string in JSON
			wantJSON: `{"id":"test-id","ts":"2024-01-01T12:00:00Z","type":"note","content":"Hello World","ctx":{}}`,
		},
		{
			name: "Event with Context",
			event: WipsEvent{
				ID:      "test-id-2",
				TS:      fixedTime,
				Type:    EventTypeGitCommit,
				Content: "Commit msg",
				Ctx: Context{
					RepoID: stringPtr("repo-1"),
					Branch: "main",
				},
			},
			wantJSON: `{"id":"test-id-2","ts":"2024-01-01T12:00:00Z","type":"git_commit","content":"Commit msg","ctx":{"repoId":"repo-1","branch":"main"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.wantJSON {
				t.Errorf("Marshal() = %s, want %s", got, tt.wantJSON)
			}

			// Round trip check
			var unmarshaled WipsEvent
			if err := json.Unmarshal(got, &unmarshaled); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}
			// Compare fields (TS might need truncated comparison if we weren't careful, but here we used pure UTC)
			if !unmarshaled.TS.Equal(tt.event.TS) {
				t.Errorf("RoundTrip TS = %v, want %v", unmarshaled.TS, tt.event.TS)
			}
			if unmarshaled.ID != tt.event.ID {
				t.Errorf("RoundTrip ID = %v, want %v", unmarshaled.ID, tt.event.ID)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
