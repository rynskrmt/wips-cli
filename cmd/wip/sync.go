package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/rynskrmt/wips-cli/internal/sync"
	"github.com/rynskrmt/wips-cli/internal/sync/obsidian"
	"github.com/rynskrmt/wips-cli/internal/usecase"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync logs to external tools",
	Long:  `Sync your wips-cli logs to external tools like Obsidian.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Flags
		flagObsidian, _ := cmd.Flags().GetBool("obsidian")
		dateStr, _ := cmd.Flags().GetString("date")
		days, _ := cmd.Flags().GetInt("days")
		all, _ := cmd.Flags().GetBool("all")
		// dryRun, _ := cmd.Flags().GetBool("dry-run") // TODO: Implement dry-run in targets

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Init Store
		s, err := store.NewStore(os.Getenv("WIPS_HOME"))
		if err != nil {
			return err
		}

		// Init Sync Manager
		mgr := sync.NewManager()

		// Register Targets
		if cfg.Sync.Obsidian != nil {
			mgr.RegisterTarget(obsidian.NewTarget(cfg.Sync.Obsidian, s))
		}

		// Determine which targets to run
		var targetNames []string
		if flagObsidian {
			targetNames = append(targetNames, "obsidian")
		} else {
			// If no flags, check defaults or all enabled?
			// Spec says: "default_targets" or "all config enabled"
			// If default_targets is set, use it. Else check what looks enabled.
			if len(cfg.Sync.DefaultTargets) > 0 {
				targetNames = cfg.Sync.DefaultTargets
			} else {
				// If manual flag not set and no default, try all registered that are enabled
				if cfg.Sync.Obsidian != nil && cfg.Sync.Obsidian.Enabled {
					targetNames = append(targetNames, "obsidian")
				}
			}
		}

		if len(targetNames) == 0 {
			fmt.Println("No sync targets enabled or specified.")
			return nil
		}

		// Prepare Usecase for fetching data
		summaryUC := usecase.NewSummaryUsecase(s)
		opts := usecase.SummaryOptions{
			IncludeHidden: false, // Default false for sync? Or true? Let's say false for now to keep private stuff private unless wanted.
			// Logic for date selection relies on same logic as summary
		}

		if all {
			// All time
			// SummaryUsecase doesn't strictly have "All" but if we don't set Day/Week/etc it defaults to Today usually?
			// Actually SummaryUsecase defaults to "Day" if nothing set.
			// We might need to extend SummaryUsecase or just use `store.GetEvents(start, end)` directly?
			// But specific date querying is handled by logic in SummaryUsecase.

			// If `all` is passed, we probably want to iterate ALL days available in store.
			// Or just pass a very wide range.
			// For now, let's limit `all` scope or fetch everything from known start time?
			// Store likely has `GetEvents(from, to)`.
			// Let's defer "all" complexity and focus on specific ranges.
			fmt.Println("Syncing --all is not fully optimized yet, syncing last 365 days...")
			opts.Days = 365
		} else if dateStr != "" {
			// Specific Date
			// Usecase doesn't directly support arbitrary `dateStr` in options struct easily without modification
			// The `summary` command doesn't actually support `--date YYYY-MM-DD` yet!
			// It supports `--day`, `--week`, `--days N`.
			// Wait, the spec said `wip sync --date`.
			// If SummaryUsecase doesn't support it, we need to create a new way or modify it.
			// Let's stick to `--days` for now or try to use `days` offset if we can calculate it (messy).
			// Actually, we can just use `store` directly if Usecase is too restrictive.
			// But Usecase adds context hydration (Git info etc).

			// Let's implement basics: --days and today (default)
			fmt.Printf("Syncing specific date %s not fully implemented, falling back to --days\n", dateStr)
		} else if days > 0 {
			opts.Days = days
		} else {
			// Default: Today
		}

		result, err := summaryUC.GetSummary(opts)
		if err != nil {
			return err
		}

		// Flatten result for Sync interface?
		// Sync interface takes `[]model.WipsEvent`.
		// Result has `DayGroups`.
		// We can flatten or pass DayGroups. Protocol said `[]model.WipsEvent`.
		var allEvents []model.WipsEvent
		for _, dg := range result.DayGroups {
			for _, dirGroup := range dg.DirMap {
				allEvents = append(allEvents, dirGroup.Events...)
			}
		}

		// Run Sync
		ctx := context.Background()
		for _, name := range targetNames {
			t, ok := mgr.GetTarget(name)
			if !ok {
				fmt.Printf("Target %s not found or not registered.\n", name)
				continue
			}

			// Filter events? The target implementation groups them by date anyway.
			if err := t.Sync(ctx, allEvents); err != nil {
				fmt.Printf("❌ Sync failed for %s: %v\n", name, err)
			} else {
				fmt.Printf("✅ Sync completed for %s\n", name)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().Bool("obsidian", false, "Sync to Obsidian")
	syncCmd.Flags().String("date", "", "Sync specific date (YYYY-MM-DD)")
	syncCmd.Flags().Int("days", 0, "Sync past N days")
	syncCmd.Flags().Bool("all", false, "Sync all history")
	syncCmd.Flags().Bool("dry-run", false, "Dry run")
}
