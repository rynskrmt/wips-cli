package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/rynskrmt/wips-cli/internal/env"
	"github.com/rynskrmt/wips-cli/internal/git"
	"github.com/rynskrmt/wips-cli/internal/id"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
)

// ContentProvider is a function that retrieves content for an event (e.g. running git show).
type ContentProvider func() ([]byte, error)

// CaptureUsecase defines the business logic for capturing external events (like git commits).
type CaptureUsecase interface {
	// CaptureEvent captures an event of a specific type.
	// It uses the contentProvider to get the event payload.
	CaptureEvent(eventType string, wd string, contentProvider ContentProvider) error
}

type captureUsecase struct {
	store store.Store
}

// NewCaptureUsecase creates a new CaptureUsecase.
func NewCaptureUsecase(s store.Store) CaptureUsecase {
	return &captureUsecase{store: s}
}

// CaptureEvent implementation.
// It currently supports "git-commit" events.
// It executes the content provider (e.g., git show), gathers context, and saves the event.
func (u *captureUsecase) CaptureEvent(eventType string, wd string, contentProvider ContentProvider) error {
	// Get Info
	if eventType != "git-commit" {
		return fmt.Errorf("unknown event type: %s", eventType)
	}

	// Gather Git Commit Info
	out, err := contentProvider()
	if err != nil {
		return fmt.Errorf("failed to get content: %w", err)
	}
	content := string(out)

	// Gather Context
	ctx := model.Context{}

	// Env
	envInfo := env.GetInfo()
	envID := id.GetHashID(envInfo.Host + envInfo.User)
	if err := u.store.SaveDict("env", envID, envInfo); err == nil {
		ctx.EnvID = &envID
	}

	// Repo
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

	// CWD
	cwdID := id.GetHashID(wd)
	if err := u.store.SaveDict("dirs", cwdID, wd); err == nil {
		ctx.CwdID = &cwdID
	}

	// Create Event
	event := &model.WipsEvent{
		ID:      id.GenerateULID(),
		TS:      time.Now(),
		Type:    model.EventTypeGitCommit,
		Content: strings.TrimSpace(content),
		Ctx:     ctx,
	}

	// Save
	if err := u.store.AppendEvent(event); err != nil {
		return err
	}

	fmt.Printf("Git commit captured: %s\n", event.ID)
	return nil
}
