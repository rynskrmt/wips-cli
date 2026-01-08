package usecase

import (
	"sort"
	"time"

	"github.com/rynskrmt/wips-cli/internal/filter"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
)

// SummaryUsecase handles business logic for the summary command.
// It retrieves events, filters them based on options, and groups them for display.
type SummaryUsecase struct {
	Store store.Store
}

// NewSummaryUsecase creates a new SummaryUsecase.
func NewSummaryUsecase(s store.Store) *SummaryUsecase {
	return &SummaryUsecase{Store: s}
}

// SummaryOptions defines filtering criteria for event summary.
type SummaryOptions struct {
	Week          bool     // Filter by current week
	LastWeek      bool     // Filter by last week
	Days          int      // Filter by past N days
	CommitsOnly   bool     // Show only git commits
	NotesOnly     bool     // Show only manual notes
	IncludeHidden bool     // Include hidden directories
	HiddenOnly    bool     // Show only hidden directories
	HiddenDirs    []string // List of hidden directory patterns from config
}

// SummaryResult holds the grouped data for display.
// Events are grouped by Day, then by Directory/Context.
type SummaryResult struct {
	Start     time.Time
	End       time.Time
	DayGroups []DayDirGroup
}

// DirGroup represents a group of events within a specific directory/repository.
type DirGroup struct {
	Name   string // Display name (e.g. "@wips-cli" or "📁 /path/to/dir")
	Events []model.WipsEvent
}

// DayDirGroup represents a group of events for a specific day.
type DayDirGroup struct {
	Date     string // YYYY-MM-DD
	DirMap   map[string]*DirGroup
	DirOrder []string // Sorted list of keys for DirMap
}

// GetSummary retrieves and organizes events based on options.
// It returns a tree-like structure (Day -> Directory -> Events) suitable for rendering.
func (u *SummaryUsecase) GetSummary(opts SummaryOptions) (*SummaryResult, error) {
	now := time.Now()
	var start time.Time
	end := now

	// Determine time range
	if opts.LastWeek {
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		thisMonday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
		start = thisMonday.AddDate(0, 0, -7)
		end = thisMonday.Add(-time.Nanosecond)
	} else if opts.Week {
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -(weekday - 1))
	} else if opts.Days > 0 {
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -opts.Days)
	} else {
		// Today
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	}

	events, err := u.Store.GetEvents(start, end)
	if err != nil {
		return nil, err
	}

	// Load Dicts for dir path lookup
	dirsDict, err := u.Store.LoadDict("dirs")
	if err != nil {
		dirsDict = make(map[string]interface{})
	}
	reposDict, err := u.Store.LoadDict("repos")
	if err != nil {
		reposDict = make(map[string]interface{})
	}

	// Filter
	var filteredEvents []model.WipsEvent
	for _, e := range events {
		if opts.CommitsOnly && e.Type != model.EventTypeGitCommit {
			continue
		}
		if opts.NotesOnly && e.Type != model.EventTypeNote {
			continue
		}

		// Hidden directory filtering
		if len(opts.HiddenDirs) > 0 {
			// Get dir path for this event
			var dirPath string
			if e.Ctx.CwdID != nil {
				if dp, ok := dirsDict[*e.Ctx.CwdID].(string); ok {
					dirPath = dp
				}
			}

			eventIsHidden := filter.IsHiddenDir(dirPath, opts.HiddenDirs)

			if opts.HiddenOnly {
				if !eventIsHidden {
					continue
				}
			} else if !opts.IncludeHidden {
				if eventIsHidden {
					continue
				}
			}
		}

		filteredEvents = append(filteredEvents, e)
	}
	events = filteredEvents

	if len(events) == 0 {
		return &SummaryResult{Start: start, End: end, DayGroups: nil}, nil
	}

	// Grouping
	var dayGroups []DayDirGroup
	dayGroupMap := make(map[string]*DayDirGroup)

	for _, e := range events {
		dateStr := e.TS.Format("2006-01-02")

		if _, exists := dayGroupMap[dateStr]; !exists {
			dayGroups = append(dayGroups, DayDirGroup{
				Date:   dateStr,
				DirMap: make(map[string]*DirGroup),
			})
			dayGroupMap[dateStr] = &dayGroups[len(dayGroups)-1]
		}
		dayGroup := dayGroupMap[dateStr]

		// Resolve Dir Name
		var dirName string
		if e.Ctx.RepoID != nil {
			if repoData, ok := reposDict[*e.Ctx.RepoID].(map[string]interface{}); ok {
				if name, ok := repoData["name"].(string); ok {
					dirName = "@" + name
				}
			}
		}
		if dirName == "" && e.Ctx.CwdID != nil {
			if dirPath, ok := dirsDict[*e.Ctx.CwdID].(string); ok {
				dirName = "📁 " + dirPath
			}
		}
		if dirName == "" {
			dirName = "(unknown)"
		}

		if _, exists := dayGroup.DirMap[dirName]; !exists {
			dayGroup.DirMap[dirName] = &DirGroup{Name: dirName, Events: []model.WipsEvent{}}
			dayGroup.DirOrder = append(dayGroup.DirOrder, dirName)
		}
		dayGroup.DirMap[dirName].Events = append(dayGroup.DirMap[dirName].Events, e)
	}

	// Sort
	for _, dg := range dayGroups {
		sort.Strings(dg.DirOrder)
	}

	return &SummaryResult{
		Start:     start,
		End:       end,
		DayGroups: dayGroups,
	}, nil
}
