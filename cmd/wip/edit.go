package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/rynskrmt/wips-cli/internal/app"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:     "edit [id]",
	Aliases: []string{"e"},
	Short:   "Edit an event",
	Long:    `Edit an event using the default editor ($EDITOR). If no ID is specified, the latest event of the current month is edited.`,
	Args:    cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		a, err := app.New()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// Get events for the current month
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		events, err := a.Store.GetEvents(start, now)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var completions []string
		for i := len(events) - 1; i >= 0; i-- { // Reverse order (newest first)
			e := events[i]
			// Limit content preview length
			preview := e.Content
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			// Format: "ID\tContent Preview"
			completions = append(completions, fmt.Sprintf("%s\t%s", e.ID, preview))
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize app with centralized dependencies
		a, err := app.New()
		if err != nil {
			return fmt.Errorf("failed to initialize app: %w", err)
		}

		var eventID string
		var targetEvent model.WipsEvent

		if len(args) > 0 {
			eventID = args[0]
			// Find event by ID
			uid, err := ulid.Parse(eventID)
			if err != nil {
				return fmt.Errorf("invalid event ID: %w", err)
			}
			ts := ulid.Time(uid.Time())

			// We use a small window around the timestamp to find the event
			// ULID has ms precision, while stored time.Now() has better precision.
			// Exact match won't work.
			events, err := a.Store.GetEvents(ts.Add(-1*time.Minute), ts.Add(1*time.Minute))
			if err != nil {
				return fmt.Errorf("failed to get events: %w", err)
			}

			found := false
			for _, e := range events {
				if e.ID == eventID {
					targetEvent = e
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("event not found: %s", eventID)
			}

		} else {
			// Get latest event of this month
			now := time.Now()
			start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			events, err := a.Store.GetEvents(start, now)
			if err != nil {
				return fmt.Errorf("failed to get events: %w", err)
			}
			if len(events) == 0 {
				return fmt.Errorf("no events found for this month")
			}
			targetEvent = events[len(events)-1]
			eventID = targetEvent.ID
		}

		// Edit content
		newContent, err := openEditor(targetEvent.Content)
		if err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}

		if newContent == targetEvent.Content {
			fmt.Println("No changes made.")
			return nil
		}

		// Update event
		err = a.Store.UpdateEvent(eventID, func(e *model.WipsEvent) error {
			e.Content = newContent
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update event %s: %w", eventID, err)
		}

		fmt.Printf("Event %s updated.\n", eventID)
		return nil
	},
}

func openEditor(content string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default to vim
	}

	tmpFile, err := ioutil.TempFile("", "wip-edit-*.md")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		return "", err
	}
	tmpFile.Close()

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run editor: %w", err)
	}

	bytes, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
