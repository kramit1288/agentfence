package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// FileRepository persists redacted audit events in a JSON file for local development.
type FileRepository struct {
	path string
	mu   sync.Mutex
}

func NewFileRepository(path string) *FileRepository {
	return &FileRepository{path: path}
}

func (r *FileRepository) Record(_ context.Context, event Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	events, err := r.load()
	if err != nil {
		return err
	}
	events = append(events, event)
	return r.save(events)
}

func (r *FileRepository) ListRecent(_ context.Context, limit int) ([]Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	events, err := r.load()
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}
	result := make([]Event, len(events))
	copy(result, events)
	return result, nil
}

func (r *FileRepository) load() ([]Event, error) {
	if r.path == "" {
		return nil, fmt.Errorf("audit file path is required")
	}
	raw, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Event{}, nil
		}
		return nil, fmt.Errorf("read audit store %q: %w", r.path, err)
	}
	if len(raw) == 0 {
		return []Event{}, nil
	}
	var events []Event
	if err := json.Unmarshal(raw, &events); err != nil {
		return nil, fmt.Errorf("decode audit store %q: %w", r.path, err)
	}
	if events == nil {
		events = []Event{}
	}
	return events, nil
}

func (r *FileRepository) save(events []Event) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("create audit store directory: %w", err)
	}
	raw, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("encode audit store: %w", err)
	}
	if err := os.WriteFile(r.path, raw, 0o600); err != nil {
		return fmt.Errorf("write audit store %q: %w", r.path, err)
	}
	return nil
}