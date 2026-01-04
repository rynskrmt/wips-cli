package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deleteCmd)
}

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete an event",
	Long:  `Delete an event. If no ID is specified, the latest event of the current month is deleted.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.NewStore(os.Getenv("WIPS_HOME"))
		if err != nil {
			return err
		}

		var eventID string

		if len(args) > 0 {
			eventID = args[0]
		} else {
			// Find latest event
			now := time.Now()
			start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			events, err := s.GetEvents(start, now)
			if err != nil {
				return err
			}
			if len(events) == 0 {
				return fmt.Errorf("no events found for this month")
			}
			eventID = events[len(events)-1].ID
			fmt.Printf("Deleting latest event: %s\n", events[len(events)-1].Content)
		}

		// Confirm? (Maybe later. Simple delete is fine for now.)

		err = s.DeleteEvent(eventID)
		if err != nil {
			return err
		}

		fmt.Printf("Event %s deleted.\n", eventID)
		return nil
	},
}
