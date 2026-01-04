package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/rynskrmt/wips-cli/internal/store"
	"github.com/rynskrmt/wips-cli/internal/ui"
	"github.com/rynskrmt/wips-cli/internal/usecase"
	"github.com/spf13/cobra"
)

func init() {
	// Register the summary command and its flags
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().Bool("week", false, "Show summary for this week")
	summaryCmd.Flags().Bool("last-week", false, "Show summary for last week")
	summaryCmd.Flags().Bool("day", false, "Show summary for today (default)")
	summaryCmd.Flags().IntP("days", "d", 0, "Show summary for past N days")
	summaryCmd.Flags().Bool("commits-only", false, "Show only git commits")
	summaryCmd.Flags().Bool("notes-only", false, "Show only manual notes")
	summaryCmd.Flags().StringP("out", "o", "", "Output file path (default stdout)")
	summaryCmd.Flags().StringP("format", "f", "pretty", "Output format (pretty, md, txt)")
	summaryCmd.Flags().Bool("include-hidden", false, "Include hidden directories in output")
	summaryCmd.Flags().Bool("hidden-only", false, "Show only hidden directories")
}

// summaryCmd represents the summary command, which aggregates and displays events.
var summaryCmd = &cobra.Command{
	Use:     "summary",
	Aliases: []string{"sum"},
	Short:   "Show summary of events",
	Long:    `Show summary of events within a specified period.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		week, _ := cmd.Flags().GetBool("week")
		lastWeek, _ := cmd.Flags().GetBool("last-week")
		days, _ := cmd.Flags().GetInt("days")
		commitsOnly, _ := cmd.Flags().GetBool("commits-only")
		notesOnly, _ := cmd.Flags().GetBool("notes-only")
		outPath, _ := cmd.Flags().GetString("out")
		format, _ := cmd.Flags().GetString("format")
		includeHidden, _ := cmd.Flags().GetBool("include-hidden")
		hiddenOnly, _ := cmd.Flags().GetBool("hidden-only")

		if outPath != "" && format == "pretty" {
			format = "md" // Default to markdown if outputting to file
		}

		// Load config for hidden directories
		cfg, _ := config.Load()
		var hiddenDirs []string
		if cfg != nil {
			hiddenDirs = cfg.HiddenDirectories
		}

		s, err := store.NewStore(os.Getenv("WIPS_HOME"))
		if err != nil {
			return err
		}

		uc := usecase.NewSummaryUsecase(s)
		opts := usecase.SummaryOptions{
			Week:          week,
			LastWeek:      lastWeek,
			Days:          days,
			CommitsOnly:   commitsOnly,
			NotesOnly:     notesOnly,
			IncludeHidden: includeHidden,
			HiddenOnly:    hiddenOnly,
			HiddenDirs:    hiddenDirs,
		}

		result, err := uc.GetSummary(opts)
		if err != nil {
			return err
		}

		if len(result.DayGroups) == 0 {
			fmt.Println("No events found.")
			return nil
		}

		var out io.Writer = os.Stdout
		if outPath != "" {
			f, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}

		renderer := ui.NewSummaryRenderer(out)

		if format == "pretty" && outPath == "" {
			renderer.RenderPretty(result)
		} else {
			if err := renderer.RenderExport(result, format); err != nil {
				return err
			}
			if outPath != "" {
				fmt.Printf("Exported to %s\n", outPath)
			}
		}

		return nil
	},
}
