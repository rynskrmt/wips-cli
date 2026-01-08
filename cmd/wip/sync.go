package main

import (
	"context"
	"fmt"

	"github.com/rynskrmt/wips-cli/internal/app"
	"github.com/rynskrmt/wips-cli/internal/model"
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

		// Initialize app with centralized dependencies
		a, err := app.New()
		if err != nil {
			return fmt.Errorf("failed to initialize app: %w", err)
		}

		cfg := a.Config

		// Init Sync Manager
		mgr := sync.NewManager()

		// Register Targets
		if cfg.Sync.Obsidian != nil {
			createMissing, _ := cmd.Flags().GetBool("create")
			opts := obsidian.TargetOptions{
				CreateMissing: createMissing,
			}
			mgr.RegisterTarget(obsidian.NewTarget(cfg.Sync.Obsidian, a.Store, opts))
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
		includeHidden, _ := cmd.Flags().GetBool("include-hidden")
		summaryUC := usecase.NewSummaryUsecase(a.Store)
		opts := usecase.SummaryOptions{
			IncludeHidden: includeHidden,
			HiddenDirs:    a.HiddenDirs(), // Apply hidden directory filter to respect user's privacy settings
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
			opts.Date = dateStr
		} else if days > 0 {
			opts.Days = days
		} else {
			// Default: Today
		}

		result, err := summaryUC.GetSummary(opts)
		if err != nil {
			return fmt.Errorf("failed to get summary for sync: %w", err)
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
	syncCmd.Flags().Bool("create", false, "Create daily note if missing")
	syncCmd.Flags().Bool("include-hidden", false, "Include hidden directories in sync")
}
