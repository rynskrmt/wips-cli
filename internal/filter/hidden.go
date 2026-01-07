// Package filter provides common filtering utilities for events and paths.
package filter

import (
	"path/filepath"
	"strings"
)

// IsHiddenDir checks if the given path is under any of the hidden directories.
// It uses platform-independent path separator handling.
func IsHiddenDir(path string, hiddenDirs []string) bool {
	for _, hiddenDir := range hiddenDirs {
		if strings.HasPrefix(path, hiddenDir) {
			// Exact match or is a subdirectory
			if path == hiddenDir || strings.HasPrefix(path, hiddenDir+string(filepath.Separator)) {
				return true
			}
		}
	}
	return false
}
