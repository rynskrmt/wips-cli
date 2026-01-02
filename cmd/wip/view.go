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
	TimeColor = color.New(color.FgCyan, color.Faint).SprintFunc()
	DateColor = color.New(color.FgHiWhite, color.Bold).SprintFunc()
)

func FormatEvent(e model.WipsEvent) (icon string, summary string) {
	summary = e.Content

	switch e.Type {
	case model.EventTypeNote:
		icon = "📝"
	case model.EventTypeGitCommit:
		icon = "🔧"
		if strings.HasPrefix(summary, "commit ") {
			lines := strings.Split(summary, "\n")
			if len(lines) > 0 {
				parts := strings.Fields(lines[0])
				hash := ""
				if len(parts) >= 2 {
					hash = parts[1][:7]
				}

				msg := ""
				for j := 1; j < len(lines); j++ {
					line := strings.TrimSpace(lines[j])
					if line != "" && !strings.HasPrefix(line, "Author:") && !strings.HasPrefix(line, "Date:") {
						msg = line
						break
					}
				}
				if hash != "" {
					summary = fmt.Sprintf("%s (%s)", msg, hash)
				} else {
					summary = msg
				}
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
	return fmt.Sprintf("%dh", int(d.Hours()))
}
