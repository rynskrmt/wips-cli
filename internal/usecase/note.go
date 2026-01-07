package usecase

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/rynskrmt/wips-cli/internal/env"
	"github.com/rynskrmt/wips-cli/internal/git"
	"github.com/rynskrmt/wips-cli/internal/id"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
)

// NoteUsecase defines the business logic for recording notes.
type NoteUsecase interface {
	// RecordNote creates and saves a new note event.
	// It automatically gathers context (environment, git repo, working directory).
	// Returns the recorded event or an error.
	RecordNote(message string, wd string) (*model.WipsEvent, error)
}

type noteUsecase struct {
	store store.Store
}

// NewNoteUsecase creates a new NoteUsecase instance.
func NewNoteUsecase(s store.Store) NoteUsecase {
	return &noteUsecase{store: s}
}

// RecordNote implementation.
// 1. Checks ignore patterns in config.
// 2. Collects environment info (user, host).
// 3. Collects git repository info if in a git repo.
// 4. Saves context dictionaries to store.
// 5. Appends the event to the store.
func (u *noteUsecase) RecordNote(message string, wd string) (*model.WipsEvent, error) {
	// Check Config
	cfg, err := config.Load()
	if err == nil {
		for _, pattern := range cfg.IgnorePatterns {
			matched, _ := filepath.Match(pattern, wd)
			if matched {
				fmt.Println("Ignored by config.")
				return nil, nil
			}
			// TODO: Implement more robust matching (e.g. support for relative paths, globstar)
		}
	}

	// Gather Context
	ctx := model.Context{}

	// Env Info
	envInfo := env.GetInfo()
	envID := id.GetHashID(envInfo.Host + envInfo.User)
	if err := u.store.SaveDict("env", envID, envInfo); err == nil {
		ctx.EnvID = &envID
	}

	// Repo Info
	if repoInfo, err := git.GetInfo(); err == nil && repoInfo.Root != "" {
		keyContent := repoInfo.Root
		if repoInfo.Remote != "" {
			keyContent = repoInfo.Remote
		}
		repoID := id.GetHashID(keyContent)

		if err := u.store.SaveDict("repos", repoID, repoInfo); err == nil {
			ctx.RepoID = &repoID
		}

		// Head info
		branch, head, err := git.GetHead()
		if err == nil {
			ctx.Branch = branch
			ctx.Head = head
		}
	}

	// CWD Info
	// Use the provided working directory
	cwdID := id.GetHashID(wd)
	if err := u.store.SaveDict("dirs", cwdID, wd); err == nil {
		ctx.CwdID = &cwdID
	}

	// Create Event
	event := &model.WipsEvent{
		ID:      id.GenerateULID(),
		TS:      time.Now(),
		Type:    model.EventTypeNote,
		Content: message,
		Ctx:     ctx,
	}

	// Save
	if err := u.store.AppendEvent(event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	return event, nil
}
