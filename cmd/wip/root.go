package main

import (
	"fmt"
	"os"

	"github.com/rynskrmt/wips-cli/internal/store"
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
	if len(args) == 0 {
		return cmd.Help()
	}

	message := args[0]

	// 1. Initialize Store
	s, err := store.NewStore(os.Getenv("WIPS_HOME"))
	if err != nil {
		return fmt.Errorf("failed to init store: %w", err)
	}
	if err := s.Prepare(); err != nil {
		return fmt.Errorf("failed to prepare store: %w", err)
	}

	// 2. Initialize Usecase
	u := usecase.NewNoteUsecase(s)

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	return u.RecordNote(message, cwd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
