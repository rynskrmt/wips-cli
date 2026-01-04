package main

import (
	"fmt"
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
		{
			name: "Git Commit (oneline)",
			event: model.WipsEvent{
				ID:      "5",
				Content: "bcf0b48 fix: something\n stats...",
				Type:    model.EventTypeGitCommit,
				TS:      time.Now(),
			},
			wantIcon:    "🔧",
			wantSummary: fmt.Sprintf("fix: something (%s)", HashColor("bcf0b48")),
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

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "Minutes",
			d:    10 * time.Minute,
			want: "10m",
		},
		{
			name: "Hours",
			d:    5 * time.Hour,
			want: "5h",
		},
		{
			name: "Days",
			d:    3 * 24 * time.Hour,
			want: "3d",
		},
		{
			name: "Weeks",
			d:    2 * 7 * 24 * time.Hour,
			want: "2w",
		},
		{
			name: "Exactly 24 hours",
			d:    24 * time.Hour,
			want: "1d",
		},
		{
			name: "Exactly 7 days",
			d:    7 * 24 * time.Hour,
			want: "1w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatDuration(tt.d); got != tt.want {
				t.Errorf("FormatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
