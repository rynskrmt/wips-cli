package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().Bool("week", false, "Show summary for this week")
	summaryCmd.Flags().Bool("last-week", false, "Show summary for last week")
	summaryCmd.Flags().Bool("day", false, "Show summary for today (default)")
	summaryCmd.Flags().IntP("days", "d", 0, "Show summary for past N days")
	summaryCmd.Flags().Bool("commits-only", false, "Show only git commits")
	summaryCmd.Flags().Bool("notes-only", false, "Show only manual notes")
	summaryCmd.Flags().StringP("out", "o", "", "Output file path (default stdout)")
	summaryCmd.Flags().StringP("format", "f", "pretty", "Output format (pretty, md, txt)")
}

var summaryCmd = &cobra.Command{
	Use:     "summary",
	Aliases: []string{"sum"},
	Short:   "Show summary of events",
	Long:    `Show summary of events within a specified period.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		week, _ := cmd.Flags().GetBool("week")
		lastWeek, _ := cmd.Flags().GetBool("last-week")
		// day, _ := cmd.Flags().GetBool("day") // default
		days, _ := cmd.Flags().GetInt("days")
		commitsOnly, _ := cmd.Flags().GetBool("commits-only")
		notesOnly, _ := cmd.Flags().GetBool("notes-only")
		outPath, _ := cmd.Flags().GetString("out")
		format, _ := cmd.Flags().GetString("format")

		if outPath != "" && format == "pretty" {
			format = "md" // Default to markdown if outputting to file
		}

		s, err := store.NewStore(os.Getenv("WIPS_HOME"))
		if err != nil {
			return err
		}

		// Load dicts for directory/repo lookups
		dirsDict, err := s.LoadDict("dirs")
		if err != nil {
			dirsDict = make(map[string]interface{})
		}
		reposDict, err := s.LoadDict("repos")
		if err != nil {
			reposDict = make(map[string]interface{})
		}

		now := time.Now()
		var start time.Time
		end := now

		// Date Logic
		if lastWeek {
			// Last Week (Mon-Sun of previous week)
			// Calculate start of this week (Monday)
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			thisMonday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
			start = thisMonday.AddDate(0, 0, -7)
			end = thisMonday.Add(-time.Nanosecond) // End of Sunday
		} else if week {
			// This Week (Mon-Now)
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
		} else if days > 0 {
			// Past N days
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -days)
		} else {
			// Today (Default)
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		}

		events, err := s.GetEvents(start, end)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		// Filter
		var filteredEvents []model.WipsEvent
		for _, e := range events {
			if commitsOnly && e.Type != model.EventTypeGitCommit {
				continue
			}
			if notesOnly && e.Type != model.EventTypeNote {
				continue
			}
			filteredEvents = append(filteredEvents, e)
		}
		events = filteredEvents

		if len(events) == 0 {
			fmt.Println("No events found after filtering.")
			return nil
		}

		// Grouping: Date -> Repo/Project
		type DirGroup struct {
			Name   string
			Events []model.WipsEvent
		}

		type DayDirGroup struct {
			Date     string
			DirMap   map[string]*DirGroup
			DirOrder []string
		}

		var dayGroups []DayDirGroup
		dayGroupMap := make(map[string]*DayDirGroup)

		for _, e := range events {
			dateStr := e.TS.Format("2006-01-02")

			// Get or create day group
			if _, exists := dayGroupMap[dateStr]; !exists {
				dayGroups = append(dayGroups, DayDirGroup{
					Date:   dateStr,
					DirMap: make(map[string]*DirGroup),
				})
				dayGroupMap[dateStr] = &dayGroups[len(dayGroups)-1]
			}
			dayGroup := dayGroupMap[dateStr]

			// Determine directory name (Repo > Dir)
			var dirName string
			if e.Ctx.RepoID != nil {
				if repoData, ok := reposDict[*e.Ctx.RepoID].(map[string]interface{}); ok {
					if name, ok := repoData["name"].(string); ok {
						dirName = "@" + name
					}
				}
			}
			if dirName == "" && e.Ctx.CwdID != nil {
				if dirPath, ok := dirsDict[*e.Ctx.CwdID].(string); ok {
					dirName = "📁 " + dirPath
				}
			}
			if dirName == "" {
				dirName = "(unknown)"
			}

			// Get or create dir group within day
			if _, exists := dayGroup.DirMap[dirName]; !exists {
				dayGroup.DirMap[dirName] = &DirGroup{Name: dirName, Events: []model.WipsEvent{}}
				dayGroup.DirOrder = append(dayGroup.DirOrder, dirName)
			}
			dayGroup.DirMap[dirName].Events = append(dayGroup.DirMap[dirName].Events, e)
		}

		// Sort dir order alphabetically for consistency
		for _, dg := range dayGroups {
			sort.Strings(dg.DirOrder)
		}

		if format == "pretty" {
			// Styles
			dateStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).MarginTop(1)
			repoStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75")).MarginLeft(2)
			timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			descStyle := lipgloss.NewStyle()

			// Setup TabWriter for aligned output within groups
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			for _, dg := range dayGroups {
				t, _ := time.Parse("2006-01-02", dg.Date)
				header := t.Format("2006-01-02 (Mon)")
				if dg.Date == time.Now().Format("2006-01-02") {
					header += " [Today]"
				}
				fmt.Println(dateStyle.Render(header))

				for _, dirName := range dg.DirOrder {
					dirGroup := dg.DirMap[dirName]
					fmt.Println(repoStyle.Render(dirGroup.Name))

					for _, e := range dirGroup.Events {
						timeStr := timeStyle.Render(e.TS.Format("15:04"))
						icon, summaryStr := FormatEvent(e)

						// Truncate summary if too long?
						// For now keep it raw but clean
						summaryStr = descStyle.Render(summaryStr)

						// Use specific format for easier reading
						// Indent 4 spaces
						fmt.Fprintf(w, "    %s\t%s  %s\t\n", timeStr, icon, summaryStr)
					}
					w.Flush() // Flush after each repo for clean blocks
				}
			}
		} else {
			// Export Logic (md, txt)
			var output strings.Builder // Need to import strings if not already there, wait... checked imports, strings is NOT imported in original file.
			// Will rely on "strings" being added or used. Wait, I should add the import in a separate tool call or include it here if I can edit imports.
			// The tool says "Multiple, non-contiguous edits". I can add import.

			if format == "md" {
				output.WriteString(fmt.Sprintf("# Activities (%s - %s)\n\n", start.Format("2006-01-02"), end.Format("2006-01-02")))
			} else {
				output.WriteString(fmt.Sprintf("Activities (%s - %s)\n\n", start.Format("2006-01-02"), end.Format("2006-01-02")))
			}

			// Iterate over pre-grouped data
			for _, dg := range dayGroups {
				dateTitle := dg.Date
				if format == "md" {
					output.WriteString(fmt.Sprintf("\n## %s\n\n", dateTitle))
				} else {
					output.WriteString(fmt.Sprintf("\n[%s]\n", dateTitle))
				}

				for _, dirName := range dg.DirOrder {
					dirGroup := dg.DirMap[dirName]
					// Section for Repo/Dir
					if format == "md" {
						output.WriteString(fmt.Sprintf("### %s\n\n", dirGroup.Name))
					} else {
						output.WriteString(fmt.Sprintf("\n%s\n", dirGroup.Name))
					}

					for _, e := range dirGroup.Events {
						timeStr := e.TS.Format("15:04")
						// _, summaryStr := FormatEvent(e) // Strip icon for plain export maybe? Or keep it? The user might want clean text.
						// Export usually implies cleaner text. Let's use clean text logic from export.go

						// Re-implement clean content extraction to avoid ANSI codes from FormatEvent if it adds any (it returns icon and string).
						// The old export.go just used e.Content.
						content := e.Content
						if e.Type == model.EventTypeGitCommit {
							lines := strings.Split(content, "\n")
							if len(lines) > 0 {
								content = lines[0]
								// Handle "hash msg" format
								parts := strings.Fields(content)
								if len(parts) >= 2 {
									// Assume first part is hash
									hash := parts[0]
									msg := strings.TrimSpace(strings.TrimPrefix(content, hash))
									content = fmt.Sprintf("%s [%s]", msg, hash)
								}
							}
						}

						if format == "md" {
							output.WriteString(fmt.Sprintf("- **%s**: %s\n", timeStr, content))
						} else {
							output.WriteString(fmt.Sprintf("- %s %s\n", timeStr, content))
						}
					}
					output.WriteString("\n")
				}
			}

			result := output.String()
			if outPath != "" {
				if err := os.WriteFile(outPath, []byte(result), 0644); err != nil {
					return err
				}
				fmt.Printf("Exported to %s\n", outPath)
			} else {
				fmt.Print(result)
			}
		}

		return nil
	},
}

// Simple legacy formatter integration
// Ideally this should be shared or imported, assuming FormatEvent is available in the package
func FormatEventLocal(e model.WipsEvent) (string, string) {
	// ... logic similar to root.go or just rely on existing if in same package
	// Assuming FormatEvent IS available in same package main (which it likely is if root.go is main)
	return FormatEvent(e)
}
