package ui

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
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
				// Use the centralized format function
				icon, summaryStr := FormatEventForSummary(e)

				summaryStr = descStyle.Render(summaryStr)
				fmt.Fprintf(w, "    %s\t%s  %s\n", timeStr, icon, summaryStr)
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
				// Use the centralized format function
				content := FormatEventPlain(e)

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
