package main

import (
	"fmt"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/spf13/cobra"
)

var configSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manage sync configuration",
	Long:  `Manage configuration for sync targets.`,
}

var configSyncObsidianCmd = &cobra.Command{
	Use:   "obsidian",
	Short: "Manage Obsidian sync configuration",
}

var configSyncObsidianEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable and configure Obsidian sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		path, _ := cmd.Flags().GetString("path")

		if cfg.Sync.Obsidian == nil {
			cfg.Sync.Obsidian = &config.ObsidianConfig{}
		}

		cfg.Sync.Obsidian.Enabled = true
		if path != "" {
			cfg.Sync.Obsidian.Path = path
		}

		// Ensure sensible defaults if fresh enable
		if cfg.Sync.Obsidian.SectionHeader == "" {
			cfg.Sync.Obsidian.SectionHeader = "## wips-cli logs"
		}
		if cfg.Sync.Obsidian.DailyFilenameFormat == "" {
			cfg.Sync.Obsidian.DailyFilenameFormat = "{{yyyy}}-{{mm}}-{{dd}}.md"
		}

		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Println("✅ Obsidian sync enabled.")
		return nil
	},
}

var configSyncObsidianDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable Obsidian sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.Sync.Obsidian != nil {
			cfg.Sync.Obsidian.Enabled = false
			if err := cfg.Save(); err != nil {
				return err
			}
		}

		fmt.Println("✅ Obsidian sync disabled.")
		return nil
	},
}

var configSyncObsidianSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set specific Obsidian config values",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if cfg.Sync.Obsidian == nil {
			return fmt.Errorf("obsidian sync is not enabled/configured. Run 'enable' first.")
		}

		path, _ := cmd.Flags().GetString("path")

		changed := false
		if path != "" {
			cfg.Sync.Obsidian.Path = path
			changed = true
		}

		if changed {
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("✅ Configuration updated.")
		} else {
			fmt.Println("No changes specified.")
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(configSyncCmd)
	configSyncCmd.AddCommand(configSyncObsidianCmd)

	configSyncObsidianCmd.AddCommand(configSyncObsidianEnableCmd)
	configSyncObsidianEnableCmd.Flags().String("path", "", "Path to Obsidian Daily Notes (e.g. ~/ObsidianVault/Daily)")
	configSyncObsidianEnableCmd.MarkFlagRequired("path")

	configSyncObsidianCmd.AddCommand(configSyncObsidianDisableCmd)

	configSyncObsidianCmd.AddCommand(configSyncObsidianSetCmd)
	configSyncObsidianSetCmd.Flags().String("path", "", "Path to Obsidian Daily Notes")
}
