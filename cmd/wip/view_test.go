package main

import (
	"testing"
	"time"

	"github.com/rynskrmt/wips-cli/internal/model"
)

func TestFormatEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       model.WipsEvent
		wantIcon    string
		wantSummary string
	}{
		{
			name: "Single line without newline",
			event: model.WipsEvent{
				ID:      "1",
				Content: "Hello World",
				Type:    model.EventTypeNote,
				TS:      time.Now(),
			},
			wantIcon:    "📝",
			wantSummary: "Hello World",
		},
		{
			name: "Single line with newline (Vim style)",
			event: model.WipsEvent{
				ID:      "2",
				Content: "Hello World\n",
				Type:    model.EventTypeNote,
				TS:      time.Now(),
			},
			wantIcon:    "📝",
			wantSummary: "Hello World",
		},
		{
			name: "Multi line",
			event: model.WipsEvent{
				ID:      "3",
				Content: "Hello\nWorld",
				Type:    model.EventTypeNote,
				TS:      time.Now(),
			},
			wantIcon:    "📝",
			wantSummary: "Hello ...",
		},
		{
			name: "Multi line with trailing newline",
			event: model.WipsEvent{
				ID:      "4",
				Content: "Hello\nWorld\n",
				Type:    model.EventTypeNote,
				TS:      time.Now(),
			},
			wantIcon:    "📝",
			wantSummary: "Hello ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIcon, gotSummary := FormatEvent(tt.event)
			if gotIcon != tt.wantIcon {
				t.Errorf("FormatEvent() gotIcon = %v, want %v", gotIcon, tt.wantIcon)
			}
			if gotSummary != tt.wantSummary {
				t.Errorf("FormatEvent() gotSummary = %q, want %q", gotSummary, tt.wantSummary)
			}
		})
	}
}
