package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rynskrmt/wips-cli/internal/app"
	"github.com/rynskrmt/wips-cli/internal/usecase"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(captureCmd)
}

var captureCmd = &cobra.Command{
	Use:    "capture [type]",
	Short:  "Capture internal events (e.g. git-commit)",
	Hidden: true, // Internal use only
	Args:   cobra.ExactArgs(1),
	RunE:   runCaptureWrapper,
}

func runCaptureWrapper(cmd *cobra.Command, args []string) error {
	eventType := args[0]
	// Initialize App
	a, err := app.New()
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	// Default git show implementation
	defaultGitShow := func() ([]byte, error) {
		return exec.Command("git", "show", "--stat", "--oneline", "--no-color", "HEAD").Output()
	}

	u := usecase.NewCaptureUsecase(a.Store)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	return u.CaptureEvent(eventType, cwd, defaultGitShow)
}
