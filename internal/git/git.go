package git

import (
	"os/exec"
	"strings"
)

// Info represents basic repository information.
type Info struct {
	Root   string
	Remote string
}

// GetInfo returns the repository root path and remote URL.
// It returns an empty Info if the current directory is not inside a git repository.
func GetInfo() (Info, error) {
	// Check if in git repo and get root path
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		// Not in a git repo or git not installed
		return Info{}, nil
	}
	root := strings.TrimSpace(string(out))

	// Get remote URL (origin)
	// Ignore error if no remote
	cmdRemote := exec.Command("git", "remote", "get-url", "origin")
	outRemote, _ := cmdRemote.Output()
	remote := strings.TrimSpace(string(outRemote))

	return Info{
		Root:   root,
		Remote: remote,
	}, nil
}

// GetHead returns the current branch name and HEAD commit hash.
func GetHead() (string, string, error) {
	// Get branch name
	cmdBranch := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	outBranch, err := cmdBranch.Output()
	if err != nil {
		return "", "", err
	}
	branch := strings.TrimSpace(string(outBranch))

	// Get HEAD hash (short)
	cmdHash := exec.Command("git", "rev-parse", "--short", "HEAD")
	outHash, err := cmdHash.Output()
	if err != nil {
		return "", "", err
	}
	hash := strings.TrimSpace(string(outHash))

	return branch, hash, nil
}
