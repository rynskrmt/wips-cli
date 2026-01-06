package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rynskrmt/wips-cli/internal/store"
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
	// Initialize Store
	s, err := store.NewStore(os.Getenv("WIPS_HOME"))
	if err != nil {
		return err
	}
	if err := s.Prepare(); err != nil {
		return err
	}

	// Default git show implementation
	defaultGitShow := func() ([]byte, error) {
		return exec.Command("git", "show", "--stat", "--oneline", "--no-color", "HEAD").Output()
	}

	u := usecase.NewCaptureUsecase(s)
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	return u.CaptureEvent(eventType, cwd, defaultGitShow)
}
