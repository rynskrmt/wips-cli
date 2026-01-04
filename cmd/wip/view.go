package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/rynskrmt/wips-cli/internal/model"
)

// ViewUtils contains shared formatting logic
var (
	TimeColor       = color.New(color.FgCyan, color.Faint).SprintFunc()
	TimeColorRecent = color.New(color.FgCyan, color.Bold).SprintFunc()
	HashColor       = color.New(color.FgYellow).SprintFunc()
	DateColor       = color.New(color.FgHiWhite, color.Bold).SprintFunc()
)

func FormatEvent(e model.WipsEvent) (icon string, summary string) {
	summary = e.Content

	switch e.Type {
	case model.EventTypeNote:
		icon = "📝"
	case model.EventTypeGitCommit:
		icon = "🔧"
		// Handle "hash msg" format (git show --oneline)
		lines := strings.Split(summary, "\n")
		if len(lines) > 0 {
			firstLine := lines[0]
			parts := strings.Fields(firstLine)

			// Check for "commit <hash>" format (standard git log)
			if len(parts) >= 2 && parts[0] == "commit" {
				// Handle standard log format if needed
			} else if len(parts) >= 2 {
				// Assume "hash msg" format
				hash := parts[0]
				// Use the rest of the line as message
				msg := strings.TrimSpace(strings.TrimPrefix(firstLine, hash))
				summary = fmt.Sprintf("%s (%s)", msg, HashColor(hash))
			}
		}

	default:
		icon = "•"
	}

	// Single line summary safety
	summary = strings.TrimSpace(summary)
	if idx := strings.Index(summary, "\n"); idx != -1 {
		summary = summary[:idx] + " ..."
	}
	return icon, summary
}

func FormatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
}
