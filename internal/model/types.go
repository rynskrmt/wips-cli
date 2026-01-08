package model

import (
	"encoding/json"
	"time"
)

// EventType represents the type of the event.
type EventType string

const (
	EventTypeNote      EventType = "note"
	EventTypeGitCommit EventType = "git_commit"
	EventTypeUndo      EventType = "undo"
)

// WipsEvent represents a single event log in the WIPS system.
// It is the core data structure stored in the daily ndjson files.
type WipsEvent struct {
	// ID is the unique identifier for the event, using ULID format.
	ID string `json:"id"`

	// TS is the timestamp of the event in ISO 8601 format.
	TS time.Time `json:"ts"`

	// Type distinguishes the kind of event (e.g., "note", "git_commit").
	Type EventType `json:"type"`

	// Content is the main payload of the event (the note text or commit message).
	Content string `json:"content"`

	// Ctx contains environmental context associated with the event (repository, cwd, etc.).
	Ctx Context `json:"ctx"`

	// Meta is a flexible field for any additional metadata, stored as raw JSON.
	Meta json.RawMessage `json:"meta,omitempty"`
}

// Context represents the environment state context for an event.
type Context struct {
	RepoID *string `json:"repoId,omitempty"` // Pointer to allow omitting
	CwdID  *string `json:"cwdId,omitempty"`
	EnvID  *string `json:"envId,omitempty"`
	Branch string  `json:"branch,omitempty"`
	Head   string  `json:"head,omitempty"`
}

// RepoInfo represents repository information stored in dict/repos.json
type RepoInfo struct {
	Name   string `json:"name"`
	Root   string `json:"root"`
	Remote string `json:"remote,omitempty"`
}

// EnvInfo represents environment information stored in dict/env.json
type EnvInfo struct {
	Host string `json:"host"`
	OS   string `json:"os"`
	Arch string `json:"arch"`
	User string `json:"user"`
}
