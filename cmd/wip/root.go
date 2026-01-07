package main

import (
	"fmt"
	"os"

	"github.com/rynskrmt/wips-cli/internal/app"
	"github.com/rynskrmt/wips-cli/internal/usecase"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wip [message]",
	Short: "CLI tool for quick memos and lightweight journaling",
	Long:  `wip is a CLI tool for developers' quick memos and lightweight journaling with automatic git commit capture.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runNoteWrapper,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func runNoteWrapper(cmd *cobra.Command, args []string) error {
	// 1. Initialize App (Centralized)
	a, err := app.New()
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	// 2. Initialize Usecase
	u := usecase.NewNoteUsecase(a.Store)

	if len(args) == 0 {
		return runInteractive(u)
	}

	message := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	event, err := u.RecordNote(message, cwd)
	if err != nil {
		return err
	}
	if event != nil {
		fmt.Printf("✅ Note recorded: %s (ID: %s)\n", event.Content, event.ID)
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
