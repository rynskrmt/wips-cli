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

// WipsEvent represents a single event log.
type WipsEvent struct {
	ID      string          `json:"id"`             // ULID
	TS      time.Time       `json:"ts"`             // ISO 8601
	Type    EventType       `json:"type"`           // "note" | "git_commit" | "undo"
	Content string          `json:"content"`        // Main text
	Ctx     Context         `json:"ctx"`            // Context info
	Meta    json.RawMessage `json:"meta,omitempty"` // Extension fields
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
