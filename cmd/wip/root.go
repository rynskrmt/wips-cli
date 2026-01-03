package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/rynskrmt/wips-cli/internal/env"
	"github.com/rynskrmt/wips-cli/internal/git"
	"github.com/rynskrmt/wips-cli/internal/id"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wip [message]",
	Short: "Local-first Event Logger",
	Long:  `wip is a local-first event logger for developers.`,
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

	return RunNote(s, message)
}

// RunNote executes the logic for the main wip command (recording a note)
func RunNote(s *store.Store, message string) error {
	// 0. Check Config
	cfg, err := config.Load()
	if err == nil {
		cwd, _ := os.Getwd()
		for _, pattern := range cfg.IgnorePatterns {
			matched, _ := filepath.Match(pattern, cwd)
			if matched {
				fmt.Println("Ignored by config.")
				return nil
			}
			// Also check if cwd is inside the ignored pattern (if pattern is a dir)
			// Simple glob matching is weak for paths.
			// Ideally we use something that understands paths.
			// But for now, let's stick to simple glob of the full path or relative?
			// Usually ignore patterns are relative to repo root, but here global config is likely absolute paths or generic names?
			// If pattern is "**/secret", we need robust matching.
			// filepath.Match "Start with" logic?
			// If user puts "/Users/foo/secret", exact match works.
			// If user puts "*/secret", it works for single level?
			// glob matching is tricky.
			// Let's assume user provides shell-style globs.
			// matching against `cwd`.
		}
	}

	// 2. Gather Context
	ctx := model.Context{}

	// 2.1 Env Info
	envInfo := env.GetInfo()
	envID := id.GetHashID(envInfo.Host + envInfo.User)
	if err := s.SaveDict("env", envID, envInfo); err == nil {
		ctx.EnvID = &envID
	}

	// 2.2 Repo Info
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

	// 2.3 CWD Info
	cwd, err := os.Getwd()
	if err == nil {
		cwdID := id.GetHashID(cwd)
		if err := s.SaveDict("dirs", cwdID, cwd); err == nil {
			ctx.CwdID = &cwdID
		}
	}

	// 3. Create Event
	event := &model.WipsEvent{
		ID:      id.GenerateULID(),
		TS:      time.Now(),
		Type:    model.EventTypeNote,
		Content: message,
		Ctx:     ctx,
	}

	// 4. Save
	if err := s.AppendEvent(event); err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	fmt.Printf("Note recorded: %s (ID: %s)\n", message, event.ID)
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
