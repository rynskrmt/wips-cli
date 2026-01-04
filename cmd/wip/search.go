package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/spf13/cobra"
	"github.com/tj/go-naturaldate"
)

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("from", "f", "", "Start date (e.g. 'yesterday', '2023-01-01')")
	searchCmd.Flags().StringP("to", "t", "", "End date (e.g. 'today', '2023-12-31')")
	searchCmd.Flags().BoolP("regex", "r", false, "Treat query as regular expression")
	searchCmd.Flags().StringSlice("tag", []string{}, "Filter by tags (e.g. 'bug', 'feature')")
	searchCmd.Flags().String("type", "", "Filter by event type (note, commit)")
}

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for events",
	Long:  `Search for events containing the specified keyword using natural language date filters.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}
		fromStr, _ := cmd.Flags().GetString("from")
		toStr, _ := cmd.Flags().GetString("to")
		isRegex, _ := cmd.Flags().GetBool("regex")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		eventType, _ := cmd.Flags().GetString("type")

		s, err := store.NewStore(os.Getenv("WIPS_HOME"))
		if err != nil {
			return err
		}

		// 1. Date Parsing
		now := time.Now()
		var start time.Time
		var end = now

		if fromStr != "" {
			parsed, err := naturaldate.Parse(fromStr, now)
			if err != nil {
				// Try standard parsing
				parsed, err = time.Parse("2006-01-02", fromStr)
				if err != nil {
					return fmt.Errorf("could not parse 'from' date: %w", err)
				}
			}
			start = parsed
		} else {
			// If no start date specified, default to beginning of 2020 (arbitrary past date to cover most history)
			start = time.Date(2020, 1, 1, 0, 0, 0, 0, time.Now().Location())
		}

		if toStr != "" {
			parsed, err := naturaldate.Parse(toStr, now)
			if err != nil {
				parsed, err = time.Parse("2006-01-02", toStr)
				if err != nil {
					return fmt.Errorf("could not parse 'to' date: %w", err)
				}
			}
			end = parsed.Add(24*time.Hour - time.Nanosecond) // End of that day
		}

		events, err := s.GetEvents(start, end)
		if err != nil {
			return err
		}

		// 2. Prepare Regex
		var re *regexp.Regexp
		if isRegex && query != "" {
			re, err = regexp.Compile(query)
			if err != nil {
				return fmt.Errorf("invalid regex: %w", err)
			}
		}

		var matchedEvents []model.WipsEvent

		for _, e := range events {
			// Filter by Type
			if eventType != "" {
				if eventType == "note" && e.Type != model.EventTypeNote {
					continue
				}
				if eventType == "commit" && e.Type != model.EventTypeGitCommit {
					continue
				}
			}

			// Filter by Content (Query)
			if query != "" {
				if isRegex {
					if !re.MatchString(e.Content) {
						continue
					}
				} else {
					if !strings.Contains(strings.ToLower(e.Content), strings.ToLower(query)) {
						continue
					}
				}
			}

			// Filter by Tags
			if len(tags) > 0 {
				hasTag := false
				for _, tag := range tags {
					// Check if content contains #tag
					tagRef := "#" + tag
					if strings.Contains(strings.ToLower(e.Content), strings.ToLower(tagRef)) {
						hasTag = true
						break
					}
				}
				if !hasTag {
					continue
				}
			}

			matchedEvents = append(matchedEvents, e)
		}

		if len(matchedEvents) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		// Sort Oldest -> Newest
		sort.Slice(matchedEvents, func(i, j int) bool {
			return matchedEvents[i].TS.Before(matchedEvents[j].TS)
		})

		// Render
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

		for _, e := range matchedEvents {
			timeStr := dateStyle.Render(e.TS.Format("2006-01-02 15:04"))
			icon, summary := FormatEvent(e)

			fmt.Fprintf(w, "%s\t%s  %s\n", timeStr, icon, summary)
		}
		w.Flush()

		return nil
	},
}
