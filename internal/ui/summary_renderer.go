package ui

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/usecase"
)

// SummaryRenderer handles rendering of summary results
type SummaryRenderer struct {
	Out io.Writer
}

func NewSummaryRenderer(out io.Writer) *SummaryRenderer {
	return &SummaryRenderer{Out: out}
}

// RenderPretty renders the summary in a pretty CLI format
func (r *SummaryRenderer) RenderPretty(result *usecase.SummaryResult) {
	dateStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).MarginTop(1)
	repoStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75")).MarginLeft(2)
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	descStyle := lipgloss.NewStyle()

	w := tabwriter.NewWriter(r.Out, 0, 0, 2, ' ', 0)

	for _, dg := range result.DayGroups {
		t, _ := time.Parse("2006-01-02", dg.Date)
		header := t.Format("2006-01-02 (Mon)")
		if dg.Date == time.Now().Format("2006-01-02") {
			header += " [Today]"
		}
		fmt.Fprintln(r.Out, dateStyle.Render(header))

		for _, dirName := range dg.DirOrder {
			dirGroup := dg.DirMap[dirName]
			fmt.Fprintln(r.Out, repoStyle.Render(dirGroup.Name))

			for _, e := range dirGroup.Events {
				timeStr := timeStyle.Render(e.TS.Format("15:04"))
				icon, summaryStr := FormatEvent(e)

				summaryStr = descStyle.Render(summaryStr)
				fmt.Fprintf(w, "    %s\t%s  %s\t\n", timeStr, icon, summaryStr)
			}
			w.Flush()
		}
	}
}

// RenderExport renders the summary in markdown or text format
func (r *SummaryRenderer) RenderExport(result *usecase.SummaryResult, format string) error {
	var output strings.Builder

	if format == "md" {
		output.WriteString(fmt.Sprintf("# Activities (%s - %s)\n\n", result.Start.Format("2006-01-02"), result.End.Format("2006-01-02")))
	} else {
		output.WriteString(fmt.Sprintf("Activities (%s - %s)\n\n", result.Start.Format("2006-01-02"), result.End.Format("2006-01-02")))
	}

	for _, dg := range result.DayGroups {
		dateTitle := dg.Date
		if format == "md" {
			output.WriteString(fmt.Sprintf("\n## %s\n\n", dateTitle))
		} else {
			output.WriteString(fmt.Sprintf("\n[%s]\n", dateTitle))
		}

		for _, dirName := range dg.DirOrder {
			dirGroup := dg.DirMap[dirName]
			if format == "md" {
				output.WriteString(fmt.Sprintf("### %s\n\n", dirGroup.Name))
			} else {
				output.WriteString(fmt.Sprintf("\n%s\n", dirGroup.Name))
			}

			for _, e := range dirGroup.Events {
				timeStr := e.TS.Format("15:04")
				content := cleanContent(e)

				if format == "md" {
					output.WriteString(fmt.Sprintf("- **%s**: %s\n", timeStr, content))
				} else {
					output.WriteString(fmt.Sprintf("- %s %s\n", timeStr, content))
				}
			}
			output.WriteString("\n")
		}
	}

	_, err := fmt.Fprint(r.Out, output.String())
	return err
}

// Helpers (Copied/Adapted from original logic)

func FormatEvent(e model.WipsEvent) (string, string) {
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
				// Basic check if it looks like a hash? length > 4?
				// tail.go/view.go assumes parts >= 2 is enough if type is git-commit
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

func cleanContent(e model.WipsEvent) string {
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
