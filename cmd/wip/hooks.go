package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	hooksCmd.AddCommand(hooksInstallCmd)
	rootCmd.AddCommand(hooksCmd)
}

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage git hooks",
}

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install wip git hooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Check if in git repo
		gitDir := ".git"
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			return fmt.Errorf("not a git repository (no .git directory found)")
		}

		// Prepare hook path
		hookPath := filepath.Join(gitDir, "hooks", "post-commit")

		// Create hook content
		script := `#!/bin/sh
# wip hook: fail silently if wip is not found or fails
if command -v wip >/dev/null 2>&1; then
  wip capture git-commit || true
fi
exit 0
`

		// Check if file exists
		// For MVP, checking if file exists and warning is safer.

		if _, err := os.Stat(hookPath); err == nil {
			// File exists
			fmt.Printf("Warning: %s already exists. Overwrite? [y/N] ", hookPath)
			var resp string
			fmt.Scanln(&resp)
			if resp != "y" && resp != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := os.WriteFile(hookPath, []byte(script), 0755); err != nil {
			return fmt.Errorf("failed to write hook: %w", err)
		}

		fmt.Printf("Installed git hook: %s\n", hookPath)
		return nil
	},
}
