package ui

import (
	"testing"
	"time"

	"github.com/rynskrmt/wips-cli/internal/model"
)

func TestFormatEventWithStyle(t *testing.T) {
	tests := []struct {
		name         string
		event        model.WipsEvent
		wantIcon     string
		wantContains string
	}{
		{
			name: "note event",
			event: model.WipsEvent{
				Type:    model.EventTypeNote,
				Content: "Test note message",
			},
			wantIcon:     "📝",
			wantContains: "Test note message",
		},
		{
			name: "git commit event",
			event: model.WipsEvent{
				Type:    model.EventTypeGitCommit,
				Content: "abc1234 Fix bug in parser",
			},
			wantIcon:     "🔧",
			wantContains: "Fix bug in parser",
		},
		{
			name: "undo event",
			event: model.WipsEvent{
				Type:    model.EventTypeUndo,
				Content: "Undid action",
			},
			wantIcon:     "↩️ ",
			wantContains: "Undid action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon, summary := FormatEventWithStyle(tt.event)
			if icon != tt.wantIcon {
				t.Errorf("FormatEventWithStyle() icon = %v, want %v", icon, tt.wantIcon)
			}
			if len(summary) == 0 {
				t.Errorf("FormatEventWithStyle() summary is empty")
			}
		})
	}
}

func TestFormatEventPlain(t *testing.T) {
	tests := []struct {
		name  string
		event model.WipsEvent
		want  string
	}{
		{
			name: "note event - unchanged",
			event: model.WipsEvent{
				Type:    model.EventTypeNote,
				Content: "Test note message",
			},
			want: "Test note message",
		},
		{
			name: "git commit event - reformatted",
			event: model.WipsEvent{
				Type:    model.EventTypeGitCommit,
				Content: "abc1234 Fix bug in parser",
			},
			want: "Fix bug in parser [abc1234]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatEventPlain(tt.event)
			if got != tt.want {
				t.Errorf("FormatEventPlain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{
			name:     "minutes",
			duration: 5 * time.Minute,
			want:     "5m",
		},
		{
			name:     "hours",
			duration: 3 * time.Hour,
			want:     "3h",
		},
		{
			name:     "days",
			duration: 2 * 24 * time.Hour,
			want:     "2d",
		},
		{
			name:     "weeks",
			duration: 2 * 7 * 24 * time.Hour,
			want:     "2w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("FormatDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
