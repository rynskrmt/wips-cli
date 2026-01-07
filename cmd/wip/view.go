package main

import (
	"time"

	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/ui"
)

// View utilities are now provided by internal/ui/format.go.
// These wrappers maintain backwards compatibility for any existing code references.

var (
	TimeColor       = ui.TimeColor
	TimeColorRecent = ui.TimeColorRecent
	HashColor       = ui.HashColor
	DateColor       = ui.DateColor
)

// FormatEvent returns an icon and summary for the given event.
// Delegates to internal/ui/format.go for the actual implementation.
func FormatEvent(e model.WipsEvent) (icon string, summary string) {
	return ui.FormatEventWithStyle(e)
}

// FormatDuration formats a duration into a human-readable string.
// Delegates to internal/ui/format.go for the actual implementation.
func FormatDuration(d time.Duration) string {
	return ui.FormatDuration(d)
}
