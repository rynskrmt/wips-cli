package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for events",
	Long:  `Search for events containing the specified keyword.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		s, err := store.NewStore(os.Getenv("WIPS_HOME"))
		if err != nil {
			return err
		}

		// Search from a reasonable start date to now
		start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Now()

		events, err := s.GetEvents(start, end)
		if err != nil {
			return err
		}

		found := false
		for _, e := range events {
			if strings.Contains(e.Content, query) {
				fmt.Printf("[%s] %s [%s] %s\n", e.TS.Format("2006-01-02 15:04"), e.ID, e.Type, e.Content)
				found = true
			}
		}

		if !found {
			fmt.Printf("No events found matching '%s'.\n", query)
		}

		return nil
	},
}
