package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rynskrmt/wips-cli/internal/env"
	"github.com/rynskrmt/wips-cli/internal/git"
	"github.com/rynskrmt/wips-cli/internal/id"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
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

	return RunCapture(s, eventType, defaultGitShow)
}

// GitShowFunc is a function that returns the output of git show
type GitShowFunc func() ([]byte, error)

// RunCapture runs the capture logic
func RunCapture(s store.Store, eventType string, gitShow GitShowFunc) error {
	// Get Info
	if eventType != "git-commit" {
		return fmt.Errorf("unknown event type: %s", eventType)
	}

	// Gather Git Commit Info
	out, err := gitShow()
	if err != nil {
		return fmt.Errorf("failed to get git show: %w", err)
	}
	content := string(out)

	// Gather Context
	ctx := model.Context{}

	// Env
	envInfo := env.GetInfo()
	envID := id.GetHashID(envInfo.Host + envInfo.User)
	if err := s.SaveDict("env", envID, envInfo); err == nil {
		ctx.EnvID = &envID
	}

	// Repo
	if repoInfo, err := git.GetInfo(); err == nil && repoInfo.Root != "" {
		keyContent := repoInfo.Root
		if repoInfo.Remote != "" {
			keyContent = repoInfo.Remote
		}
		repoID := id.GetHashID(keyContent)
		if err := s.SaveDict("repos", repoID, repoInfo); err == nil {
			ctx.RepoID = &repoID
		}

		// Head info
		branch, head, err := git.GetHead()
		if err == nil {
			ctx.Branch = branch
			ctx.Head = head
		}
	}

	// CWD
	cwd, err := os.Getwd()
	if err == nil {
		cwdID := id.GetHashID(cwd)
		if err := s.SaveDict("dirs", cwdID, cwd); err == nil {
			ctx.CwdID = &cwdID
		}
	}

	// Create Event
	event := &model.WipsEvent{
		ID:      id.GenerateULID(),
		TS:      time.Now(),
		Type:    model.EventTypeGitCommit,
		Content: strings.TrimSpace(content),
		Ctx:     ctx,
	}

	// 5. Save
	if err := s.AppendEvent(event); err != nil {
		return err
	}

	fmt.Printf("Git commit captured: %s\n", event.ID)
	return nil
}
