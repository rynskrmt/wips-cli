// Package ui provides user interface utilities for displaying events and summaries.
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/rynskrmt/wips-cli/internal/model"
)

// Style definitions for terminal output
var (
	TimeColor       = color.New(color.FgCyan, color.Faint).SprintFunc()
	TimeColorRecent = color.New(color.FgCyan, color.Bold).SprintFunc()
	HashColor       = color.New(color.FgYellow).SprintFunc()
	DateColor       = color.New(color.FgHiWhite, color.Bold).SprintFunc()
)

// FormatEventWithStyle returns an icon and formatted summary for the given event.
// It applies terminal color styling suitable for CLI output.
func FormatEventWithStyle(e model.WipsEvent) (icon string, summary string) {
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
	case model.EventTypeUndo:
		icon = "↩️ "
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

// FormatEventPlain returns a plain text summary suitable for markdown or file export.
// It does not include any color codes.
func FormatEventPlain(e model.WipsEvent) string {
	content := e.Content
	if e.Type == model.EventTypeGitCommit {
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			content = lines[0]
			parts := strings.Fields(content)
			if len(parts) >= 2 {
				hash := parts[0]
				msg := strings.TrimSpace(strings.TrimPrefix(content, hash))
				content = fmt.Sprintf("%s [%s]", msg, hash)
			}
		}
	}
	return content
}

// FormatEventForSummary returns an icon and summary with lipgloss styling.
// Used for the summary command output.
func FormatEventForSummary(e model.WipsEvent) (string, string) {
	icon := "📝"
	summary := e.Content
	hashStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow

	switch e.Type {
	case model.EventTypeGitCommit:
		icon = "🔧"
		lines := strings.Split(summary, "\n")
		if len(lines) > 0 {
			firstLine := lines[0]
			parts := strings.Fields(firstLine)
			// Check for "hash msg" format
			if len(parts) >= 2 {
				hash := parts[0]
				msg := strings.TrimSpace(strings.TrimPrefix(firstLine, hash))
				summary = fmt.Sprintf("%s (%s)", msg, hashStyle.Render(hash))
			} else {
				summary = firstLine
			}
		}
	case model.EventTypeUndo:
		icon = "↩️ "
	}

	// Truncate content for display (handle multiline if not git commit with special parsing)
	if e.Type != model.EventTypeGitCommit {
		lines := strings.Split(summary, "\n")
		summary = lines[0]
		if len(lines) > 1 {
			summary += "..."
		}
	}

	return icon, summary
}

// FormatDuration formats a duration into a human-readable relative time string.
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

// FormatTimeRelative formats a time as relative with appropriate color.
func FormatTimeRelative(t time.Time) string {
	d := time.Since(t)
	timeStr := FormatDuration(d)
	if d < 24*time.Hour {
		return TimeColorRecent(timeStr)
	}
	return TimeColor(timeStr)
}
