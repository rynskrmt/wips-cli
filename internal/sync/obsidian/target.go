package obsidian

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rynskrmt/wips-cli/internal/config"
	"github.com/rynskrmt/wips-cli/internal/model"
	"github.com/rynskrmt/wips-cli/internal/store"
)

type Target struct {
	cfg   *config.ObsidianConfig
	store store.Store
	opts  TargetOptions
}

type TargetOptions struct {
	CreateMissing bool
}

func NewTarget(cfg *config.ObsidianConfig, s store.Store, opts TargetOptions) *Target {
	return &Target{cfg: cfg, store: s, opts: opts}
}

func (t *Target) Name() string {
	return "obsidian"
}

func (t *Target) Sync(ctx context.Context, events []model.WipsEvent) error {
	if !t.cfg.Enabled {
		return nil
	}

	// Group events by date as Obsidian files are usually daily
	eventsByDate := make(map[string][]model.WipsEvent)
	for _, event := range events {
		dateStr := event.TS.Format("2006-01-02")
		eventsByDate[dateStr] = append(eventsByDate[dateStr], event)
	}

	for dateStr, dateEvents := range eventsByDate {
		if err := t.syncDate(dateStr, dateEvents); err != nil {
			return fmt.Errorf("failed to sync date %s: %w", dateStr, err)
		}
	}

	return nil
}

func (t *Target) syncDate(dateStr string, events []model.WipsEvent) error {
	// 1. Determine file path
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err
	}

	// Default format if not specified
	filenameFormat := t.cfg.DailyFilenameFormat
	if filenameFormat == "" {
		filenameFormat = "{{yyyy}}-{{mm}}-{{dd}}.md"
	}

	// Expand home directory in path
	targetPath := t.cfg.Path
	if strings.HasPrefix(targetPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		targetPath = filepath.Join(home, targetPath[2:])
	}

	filename := formatFilename(filenameFormat, date)
	fullPath := filepath.Join(targetPath, filename)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// 2. Generate content
	content, err := t.generateContent(date, events)
	if err != nil {
		return err
	}

	// 3. Read existing file
	existingContent := ""
	if _, err := os.Stat(fullPath); err == nil {
		b, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read existing file: %w", err)
		}
		existingContent = string(b)
	} else if os.IsNotExist(err) {
		if !t.opts.CreateMissing {
			fmt.Printf("⚠️  %s: Daily note does not exist (skipping)\n", dateStr)
			return nil
		}
		// Proceed to create
	} else {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// 4. Update file content
	newFileContent := t.updateFileContent(existingContent, content)

	// 5. Write file
	if err := os.WriteFile(fullPath, []byte(newFileContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Synced to Obsidian: %s (Events: %d)\n", fullPath, len(events))
	return nil
}

func (t *Target) generateContent(date time.Time, events []model.WipsEvent) (string, error) {
	// Load context dictionaries
	dirsDict, err := t.store.LoadDict("dirs")
	if err != nil {
		dirsDict = make(map[string]interface{})
	}
	reposDict, err := t.store.LoadDict("repos")
	if err != nil {
		reposDict = make(map[string]interface{})
	}

	var sb strings.Builder

	sectionHeader := t.cfg.SectionHeader
	if sectionHeader == "" {
		sectionHeader = "## wips-cli logs"
	}
	sb.WriteString(sectionHeader + "\n\n")

	// Group events by context (Repo or Dir)
	type DirGroup struct {
		Name   string
		Events []model.WipsEvent
	}

	dirGroups := make(map[string]*DirGroup)
	var dirOrder []string

	for _, e := range events {
		// Resolve Dir Name (Similar to usecase/summary.go)
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

		if _, exists := dirGroups[dirName]; !exists {
			dirGroups[dirName] = &DirGroup{Name: dirName, Events: []model.WipsEvent{}}
			dirOrder = append(dirOrder, dirName)
		}
		dirGroups[dirName].Events = append(dirGroups[dirName].Events, e)
	}

	sort.Strings(dirOrder)

	for _, dirName := range dirOrder {
		group := dirGroups[dirName]
		sb.WriteString(fmt.Sprintf("### %s\n\n", group.Name))

		for _, e := range group.Events {
			timeStr := e.TS.Format("15:04")
			content := cleanContent(e)
			// Using standard markdown list format
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", timeStr, content))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func (t *Target) updateFileContent(existing, newSection string) string {
	sectionHeader := t.cfg.SectionHeader
	if sectionHeader == "" {
		sectionHeader = "## wips-cli logs"
	}

	targetLevel := getHeaderLevel(sectionHeader)

	lines := strings.Split(existing, "\n")
	startIdx := -1
	endIdx := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			startIdx = i
			continue
		}
		if startIdx != -1 {
			// If we found start, look for next header
			if strings.HasPrefix(line, "#") {
				level := getHeaderLevel(line)
				if level <= targetLevel {
					endIdx = i
					break
				}
			}
		}
	}

	if startIdx == -1 {
		// Section not found, append
		if existing == "" {
			return newSection
		}

		appendAt := t.cfg.AppendAt
		if appendAt == "top" {
			// Ideally after YAML frontmatter if exists
			// Simple implementation: just prepend
			return newSection + "\n" + existing
		}
		// Default bottom
		if !strings.HasSuffix(existing, "\n") {
			existing += "\n"
		}
		return existing + "\n" + newSection
	}

	// Section found, replace
	if endIdx == -1 {
		// It was the last section
		endIdx = len(lines)
	}

	// Reconstruct
	// Keep lines before startIdx
	// Insert newSection
	// Keep lines from endIdx

	// Handle whitespace around section?
	newSection = strings.TrimRight(newSection, "\n")

	var sb strings.Builder
	for i := 0; i < startIdx; i++ {
		sb.WriteString(lines[i] + "\n")
	}
	sb.WriteString(newSection + "\n")
	if endIdx < len(lines) {
		sb.WriteString("\n")
	}
	for i := endIdx; i < len(lines); i++ {
		sb.WriteString(lines[i] + "\n")
	}

	return sb.String()
}

func getHeaderLevel(line string) int {
	trimmed := strings.TrimSpace(line)
	level := 0
	for _, c := range trimmed {
		if c == '#' {
			level++
		} else {
			break
		}
	}
	// If no hash or just text, level might be 0.
	// But in markdown, headers start with #.
	// If it doesn't start with #, level is 0.
	return level
}

func cleanContent(e model.WipsEvent) string {
	// Simple duplicate of ui.cleanContent
	content := e.Content
	if e.Type == model.EventTypeGitCommit {
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			content = lines[0]
			parts := strings.Fields(content)
			if len(parts) >= 2 {
				hash := parts[0]
				msg := strings.TrimSpace(strings.TrimPrefix(content, hash))
				content = fmt.Sprintf("%s [%s]", msg, hash)
			}
		}
	}
	return content
}

func formatFilename(format string, date time.Time) string {
	r := strings.NewReplacer(
		"{{yyyy}}", date.Format("2006"),
		"{{mm}}", date.Format("01"),
		"{{dd}}", date.Format("02"),
	)
	return r.Replace(format)
}
