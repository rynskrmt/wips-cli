package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/oklog/ulid/v2"
	"github.com/rynskrmt/wips-cli/internal/model"
)

// Store defines the interface for data persistence
type Store interface {
	Prepare() error
	AppendEvent(event *model.WipsEvent) error
	SaveDict(dictName string, key string, value interface{}) error
	LoadDict(dictName string) (map[string]interface{}, error)
	GetEvents(start, end time.Time) ([]model.WipsEvent, error)
	UpdateEvent(id string, mutator func(*model.WipsEvent) error) error
	DeleteEvent(id string) error
	GetRootDir() string
}

// FileStore handles file system operations for wips-cli
type FileStore struct {
	RootDir string
	mu      sync.Mutex // Process-internal lock, file lock used for inter-process
}

func (s *FileStore) GetRootDir() string {
	return s.RootDir
}

// NewStore creates a new store instance.
// If rootDir is empty, it attempts to find the default data directory.
func NewStore(rootDir string) (Store, error) {
	if rootDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home dir: %w", err)
		}
		// Default path: ~/Library/Application Support/wip on macOS
		// but spec says standard app data.
		configDir, err := os.UserConfigDir()
		if err != nil {
			rootDir = filepath.Join(home, ".wip")
		} else {
			rootDir = filepath.Join(configDir, "wip")
		}
	}

	return &FileStore{RootDir: rootDir}, nil
}

// Prepare ensures necessary directories exist.
func (s *FileStore) Prepare() error {
	dirs := []string{
		filepath.Join(s.RootDir, "events"),
		filepath.Join(s.RootDir, "dict"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", d, err)
		}
	}
	return nil
}

// AppendEvent appends an event to events/YYYY-MM.ndjson
func (s *FileStore) AppendEvent(event *model.WipsEvent) error {
	filename := event.TS.Format("2006-01") + ".ndjson"
	path := filepath.Join(s.RootDir, "events", filename)

	fileLock := flock.New(path)
	locked, err := fileLock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to lock event file: %w", err)
	}
	if !locked {
		return fmt.Errorf("failed to lock event file: file busy")
	}
	defer fileLock.Unlock()

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open event file: %w", err)
	}
	defer f.Close()

	bytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := f.Write(append(bytes, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// SaveDict updates a dictionary file idempotently.
// It reads the existing JSON, checks if the key exists, adds it if not, and writes back.
// Uses file locking to prevent race conditions.
func (s *FileStore) SaveDict(dictName string, key string, value interface{}) error {
	path := filepath.Join(s.RootDir, "dict", dictName+".json")

	fileLock := flock.New(path)
	locked, err := fileLock.TryLock()
	if err != nil {
		return fmt.Errorf("failed to lock dict file: %w", err)
	}
	if !locked {
		if err := fileLock.Lock(); err != nil {
			return fmt.Errorf("failed to lock dict file: %w", err)
		}
	}
	defer fileLock.Unlock()

	// Open file with RDWR to allow reading and writing
	// O_CREAT ensure it exists
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open dict file: %w", err)
	}
	defer f.Close()

	// Read content
	var content map[string]interface{}
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if fi.Size() > 0 {
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&content); err != nil {
			// If JSON is corrupted, we might want to backup and reset, but for MVP just error
			return fmt.Errorf("failed to decode dict file: %w", err)
		}
	} else {
		content = make(map[string]interface{})
	}

	// Check idempotency (simple key existence check)
	if _, exists := content[key]; exists {
		return nil // Already exists, do nothing
	}

	// Update map
	content[key] = value

	// Rewind and write
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(content); err != nil {
		return fmt.Errorf("failed to encode dict file: %w", err)
	}

	return nil
}

// GetEvents returns events within the given time range.
func (s *FileStore) GetEvents(start, end time.Time) ([]model.WipsEvent, error) {
	var events []model.WipsEvent

	// Iterate over months from start to end
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	endMonth := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, end.Location())

	for !current.After(endMonth) {
		filename := current.Format("2006-01") + ".ndjson"
		path := filepath.Join(s.RootDir, "events", filename)

		// It's possible the file doesn't exist if no events were recorded that month
		if _, err := os.Stat(path); os.IsNotExist(err) {
			current = current.AddDate(0, 1, 0)
			continue
		}

		fileEvents, err := s.readEventsFromFile(path)
		if err != nil {
			return nil, err
		}

		// Filter by exact time range
		for _, e := range fileEvents {
			if (e.TS.Equal(start) || e.TS.After(start)) && (e.TS.Equal(end) || e.TS.Before(end)) {
				events = append(events, e)
			}
		}

		current = current.AddDate(0, 1, 0)
	}

	return events, nil
}

// UpdateEvent finds an event by ID and updates it using the mutator function.
func (s *FileStore) UpdateEvent(id string, mutator func(*model.WipsEvent) error) error {
	// Parse ULID to get timestamp
	uid, err := ulid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid event ID: %w", err)
	}

	ts := ulid.Time(uid.Time())

	filename := ts.Format("2006-01") + ".ndjson"
	path := filepath.Join(s.RootDir, "events", filename)

	return s.rewriteFile(path, func(events []model.WipsEvent) ([]model.WipsEvent, error) {
		found := false
		for i := range events {
			if events[i].ID == id {
				if err := mutator(&events[i]); err != nil {
					return nil, err
				}
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("event not found: %s", id)
		}
		return events, nil
	})
}

// DeleteEvent deletes an event by ID.
func (s *FileStore) DeleteEvent(id string) error {
	// Parse ULID to get timestamp
	uid, err := ulid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid event ID: %w", err)
	}

	ts := ulid.Time(uid.Time())

	filename := ts.Format("2006-01") + ".ndjson"
	path := filepath.Join(s.RootDir, "events", filename)

	return s.rewriteFile(path, func(events []model.WipsEvent) ([]model.WipsEvent, error) {
		newEvents := make([]model.WipsEvent, 0, len(events))
		found := false
		for _, e := range events {
			if e.ID == id {
				found = true
				continue
			}
			newEvents = append(newEvents, e)
		}
		if !found {
			return nil, fmt.Errorf("event not found: %s", id)
		}
		return newEvents, nil
	})
}

// readEventsFromFile reads all events from a given file path.
func (s *FileStore) readEventsFromFile(path string) ([]model.WipsEvent, error) {
	fileLock := flock.New(path)
	// Try shared lock for reading
	locked, err := fileLock.TryRLock()
	if err != nil {
		return nil, fmt.Errorf("failed to lock file %s: %w", path, err)
	}
	if !locked {
		// Fallback to blocking lock
		if err := fileLock.RLock(); err != nil {
			return nil, fmt.Errorf("failed to lock file %s: %w", path, err)
		}
	}
	defer fileLock.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	var events []model.WipsEvent
	decoder := json.NewDecoder(f)
	for decoder.More() {
		var e model.WipsEvent
		if err := decoder.Decode(&e); err != nil {
			return nil, fmt.Errorf("failed to decode event in %s: %w", path, err)
		}
		events = append(events, e)
	}
	return events, nil
}

// rewriteFile safely rewrites a file by reading all content, applying a transformation, and writing back.
func (s *FileStore) rewriteFile(path string, transform func([]model.WipsEvent) ([]model.WipsEvent, error)) error {
	fileLock := flock.New(path)
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("failed to lock file %s: %w", path, err)
	}
	defer fileLock.Unlock()

	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer f.Close()

	var events []model.WipsEvent
	decoder := json.NewDecoder(f)
	for decoder.More() {
		var e model.WipsEvent
		if err := decoder.Decode(&e); err != nil {
			return fmt.Errorf("failed to decode event in %s: %w", path, err)
		}
		events = append(events, e)
	}

	newEvents, err := transform(events)
	if err != nil {
		return err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if err := f.Truncate(0); err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	for _, e := range newEvents {
		if err := encoder.Encode(e); err != nil {
			return fmt.Errorf("failed to encode event: %w", err)
		}
	}

	return nil
}

// LoadDict loads a dictionary file.
func (s *FileStore) LoadDict(dictName string) (map[string]interface{}, error) {
	path := filepath.Join(s.RootDir, "dict", dictName+".json")

	fileLock := flock.New(path)
	if err := fileLock.RLock(); err != nil {
		return nil, fmt.Errorf("failed to lock dict file: %w", err)
	}
	defer fileLock.Unlock()

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return make(map[string]interface{}), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open dict file: %w", err)
	}
	defer f.Close()

	var content map[string]interface{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&content); err != nil {
		if err.Error() == "EOF" {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to decode dict file: %w", err)
	}

	return content, nil
}
