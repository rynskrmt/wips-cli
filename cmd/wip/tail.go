package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rynskrmt/wips-cli/internal/app"
	"github.com/rynskrmt/wips-cli/internal/filter"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tailCmd)
	tailCmd.Flags().IntP("lines", "n", 10, "number of lines to show")
	tailCmd.Flags().BoolP("id", "i", false, "display event IDs")
	tailCmd.Flags().BoolP("global", "g", false, "show all events regardless of context")
	tailCmd.Flags().Bool("include-hidden", false, "Include hidden directories in output")
}

var tailCmd = &cobra.Command{
	Use:     "tail",
	Aliases: []string{"t"},
	Short:   "Show recent events",
	RunE: func(cmd *cobra.Command, args []string) error {
		n, _ := cmd.Flags().GetInt("lines")
		showID, _ := cmd.Flags().GetBool("id")
		global, _ := cmd.Flags().GetBool("global")
		includeHidden, _ := cmd.Flags().GetBool("include-hidden")

		// Get current CWD to filter by path hierarchy
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Initialize app with centralized dependencies
		a, err := app.New()
		if err != nil {
			return fmt.Errorf("failed to initialize app: %w", err)
		}

		// Load dirs dict for path lookup
		dirsDict, err := a.Store.LoadDict("dirs")
		if err != nil {
			dirsDict = make(map[string]interface{})
		}

		// Load repos dict
		reposDict, err := a.Store.LoadDict("repos")
		if err != nil {
			reposDict = make(map[string]interface{})
		}

		// Read current month file
		filename := time.Now().Format("2006-01") + ".ndjson"
		path := filepath.Join(a.Store.GetRootDir(), "events", filename)

		f, err := os.Open(path)
		if os.IsNotExist(err) {
			fmt.Println("No events found for this month.")
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to open events file: %w", err)
		}
		defer f.Close()

		var events []model.WipsEvent
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var e model.WipsEvent
			if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
				// Get dir path for filtering
				var dirPath string
				if e.Ctx.CwdID != nil {
					if dp, ok := dirsDict[*e.Ctx.CwdID].(string); ok {
						dirPath = dp
					}
				}

				// Hidden directory filtering using shared filter package
				if !includeHidden && filter.IsHiddenDir(dirPath, a.HiddenDirs()) {
					continue
				}

				// Filter by context if not global
				if !global {
					shouldShow := false

					if dirPath != "" && strings.HasPrefix(dirPath, cwd) {
						shouldShow = true
					}

					if !shouldShow {
						continue
					}
				}
				events = append(events, e)
			}
		}

		// Reverse back and trim
		// Take last n events
		start := 0
		if len(events) > n {
			start = len(events) - n
		}
		shownEvents := events[start:]

		// Print in chronological order (Oldest -> Newest)

		// Setup tabwriter
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		for _, e := range shownEvents {

			// Format Time using shared format package
			timeStr := ui.FormatTimeRelative(e.TS)

			icon, summary := ui.FormatEventWithStyle(e)

			// Build context string for global mode
			var ctxStr string
			if global {
				if e.Ctx.RepoID != nil {
					if repoData, ok := reposDict[*e.Ctx.RepoID].(map[string]interface{}); ok {
						if name, ok := repoData["name"].(string); ok {
							ctxStr = "@" + name
						}
					}
				}
				if ctxStr == "" && e.Ctx.CwdID != nil {
					if dirPath, ok := dirsDict[*e.Ctx.CwdID].(string); ok {
						// Use last path component
						ctxStr = "📁 " + filepath.Base(dirPath)
					}
				}
			}

			if showID {
				if ctxStr != "" {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n", timeStr, icon, summary, ctxStr, e.ID)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", timeStr, icon, summary, e.ID)
				}
			} else {
				if ctxStr != "" {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", timeStr, icon, summary, ctxStr)
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t\n", timeStr, icon, summary)
				}
			}
		}
		w.Flush()

		return nil
	},
}
