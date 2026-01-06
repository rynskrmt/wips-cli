package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configAddHiddenCmd)
	configCmd.AddCommand(configRemoveHiddenCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configEditCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage wip configuration",
	Long:  `View and modify wip configuration settings.`,
}

var configAddHiddenCmd = &cobra.Command{
	Use:   "add-hidden <path>",
	Short: "Add a directory to the hidden list",
	Long: `Add a directory to the hidden list.
Events from hidden directories will be excluded from summary and tail by default.
Use --include-hidden flag to show them.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.AddHiddenDir(args[0]); err != nil {
			return err
		}

		fmt.Printf("Added hidden directory: %s\n", args[0])
		return nil
	},
}

var configRemoveHiddenCmd = &cobra.Command{
	Use:   "remove-hidden <path>",
	Short: "Remove a directory from the hidden list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.RemoveHiddenDir(args[0]); err != nil {
			return err
		}

		fmt.Printf("Removed hidden directory: %s\n", args[0])
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		fmt.Println("=== wip configuration ===")
		fmt.Println()

		fmt.Println("Hidden Directories:")
		if len(cfg.HiddenDirectories) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, d := range cfg.HiddenDirectories {
				fmt.Printf("  - %s\n", d)
			}
		}
		fmt.Println()

		fmt.Println("Ignore Patterns:")
		if len(cfg.IgnorePatterns) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, p := range cfg.IgnorePatterns {
				fmt.Printf("  - %s\n", p)
			}
		}
		fmt.Println()

		fmt.Println("Sync Configuration:")
		if cfg.Sync.Obsidian != nil && cfg.Sync.Obsidian.Enabled {
			fmt.Printf("  Obsidian: Enabled\n")
			fmt.Printf("    Path: %s\n", cfg.Sync.Obsidian.Path)
		} else if cfg.Sync.Obsidian != nil {
			fmt.Printf("  Obsidian: Disabled (Path: %s)\n", cfg.Sync.Obsidian.Path)
		} else {
			fmt.Println("  Obsidian: Not configured")
		}

		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file in $EDITOR",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetConfigPath()
		if err != nil {
			return err
		}

		// Ensure config file exists
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if err := cfg.Save(); err != nil {
			return err
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}

		editorCmd := exec.Command(editor, configPath)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		return editorCmd.Run()
	},
}
