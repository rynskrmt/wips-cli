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

	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tailCmd)
	tailCmd.Flags().IntP("lines", "n", 10, "number of lines to show")
	tailCmd.Flags().BoolP("id", "i", false, "display event IDs")
	tailCmd.Flags().BoolP("global", "g", false, "show all events regardless of context")
}

var tailCmd = &cobra.Command{
	Use:     "tail",
	Aliases: []string{"t"},
	Short:   "Show recent events",
	RunE: func(cmd *cobra.Command, args []string) error {
		n, _ := cmd.Flags().GetInt("lines")
		showID, _ := cmd.Flags().GetBool("id")
		global, _ := cmd.Flags().GetBool("global")

		// Get current CWD to filter by path hierarchy
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		s, err := store.NewStore(os.Getenv("WIPS_HOME"))
		if err != nil {
			return err
		}

		// Load dirs dict for path lookup
		dirsDict, err := s.LoadDict("dirs")
		if err != nil {
			dirsDict = make(map[string]interface{})
		}

		// Load repos dict for context display in global mode
		reposDict, err := s.LoadDict("repos")
		if err != nil {
			reposDict = make(map[string]interface{})
		}

		// Read current month file for MVP
		filename := time.Now().Format("2006-01") + ".ndjson"
		path := filepath.Join(s.RootDir, "events", filename)

		f, err := os.Open(path)
		if os.IsNotExist(err) {
			fmt.Println("No events found for this month.")
			return nil
		}
		if err != nil {
			return err
		}
		defer f.Close()

		var events []model.WipsEvent
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var e model.WipsEvent
			if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
				// Filter by context if not global
				if !global {
					shouldShow := false

					if e.Ctx.CwdID != nil {
						if dirPath, ok := dirsDict[*e.Ctx.CwdID].(string); ok {
							// Check if dirPath starts with cwd
							if strings.HasPrefix(dirPath, cwd) {
								shouldShow = true
							}
						}
					}
					// Also include events if current dir is distinct but part of same repo?
					// For now stick to strict directory filter as requested.

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

			// Format Time
			d := time.Since(e.TS)
			timeStr := FormatDuration(d)
			if d < 24*time.Hour {
				timeStr = TimeColorRecent(timeStr)
			} else {
				timeStr = TimeColor(timeStr)
			}

			icon, summary := FormatEvent(e)

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
