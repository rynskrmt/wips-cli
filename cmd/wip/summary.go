package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().String("since", "today", "Show summary since when (today, yesterday, week, or duration like 24h)")
	summaryCmd.Flags().Bool("by-dir", false, "Group events by directory instead of date")
}

var summaryCmd = &cobra.Command{
	Use:     "summary",
	Aliases: []string{"sum"},
	Short:   "Show summary of events",
	Long:    `Show summary of events within a specified period.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		since, _ := cmd.Flags().GetString("since")
		byDir, _ := cmd.Flags().GetBool("by-dir")

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

		switch since {
		case "today":
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		case "yesterday":
			yesterday := now.AddDate(0, 0, -1)
			start = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
			end = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		case "week":
			// Start of this week (Sunday or Monday? Let's assume Monday as start of week, or just last 7 days?)
			// "this week" usually means from Monday? Or simply last 7 days.
			// Let's implement "last 7 days" for simplicity if user says "week".
			// Or maybe "this week" from Sunday.
			// Let's go with "start of this week (Sunday)".
			weekday := int(now.Weekday())
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -weekday)
		default:
			// Try parsing as duration
			d, err := time.ParseDuration(since)
			if err != nil {
				return fmt.Errorf("invalid duration format: %s", since)
			}
			start = now.Add(-d)
		}

		events, err := s.GetEvents(start, end)
		if err != nil {
			return err
		}

		if len(events) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		if len(events) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		if byDir {
			// Group by date first, then by directory within each date
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

				// Determine directory name
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

			// Print grouped output
			for _, dg := range dayGroups {
				// Print Date Header
				t, _ := time.Parse("2006-01-02", dg.Date)
				header := t.Format("2006-01-02 (Mon)")
				if dg.Date == time.Now().Format("2006-01-02") {
					header += " [Today]"
				}
				fmt.Fprintf(w, "\n%s\n", DateColor(header))

				for _, dirName := range dg.DirOrder {
					dirGroup := dg.DirMap[dirName]
					fmt.Fprintf(w, "  %s\n", dirGroup.Name)

					for _, e := range dirGroup.Events {
						timeStr := TimeColor(e.TS.Format("15:04"))
						icon, summaryStr := FormatEvent(e)
						fmt.Fprintf(w, "    %s\t%s\t%s\t\n", timeStr, icon, summaryStr)
					}
				}
			}
		} else {
			// Group events by day (original behavior)
			type DayGroup struct {
				Date   string
				Events []model.WipsEvent
			}

			var groups []DayGroup
			var currentGroup *DayGroup

			for _, e := range events {
				dateStr := e.TS.Format("2006-01-02")

				if currentGroup == nil || currentGroup.Date != dateStr {
					groups = append(groups, DayGroup{Date: dateStr, Events: []model.WipsEvent{}})
					currentGroup = &groups[len(groups)-1]
				}
				currentGroup.Events = append(currentGroup.Events, e)
			}

			for _, g := range groups {
				// Print Date Header
				t, _ := time.Parse("2006-01-02", g.Date)
				header := t.Format("2006-01-02 (Mon)")
				if g.Date == time.Now().Format("2006-01-02") {
					header += " [Today]"
				}
				fmt.Fprintf(w, "\n%s\n", DateColor(header))

				for _, e := range g.Events {
					timeStr := TimeColor(e.TS.Format("15:04"))
					icon, summaryStr := FormatEvent(e)
					fmt.Fprintf(w, "%s\t%s\t%s\t\n", timeStr, icon, summaryStr)
				}
			}
		}
		w.Flush()

		return nil
	},
}
