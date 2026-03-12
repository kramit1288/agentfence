package approval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileRepository persists approval requests in a JSON file.
type FileRepository struct {
	path string
	mu   sync.Mutex
}

func NewFileRepository(path string) *FileRepository {
	return &FileRepository{path: path}
}

func (r *FileRepository) Create(_ context.Context, request Request) (Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	requests, err := r.load()
	if err != nil {
		return Request{}, err
	}
	requests[request.ID] = request
	if err := r.save(requests); err != nil {
		return Request{}, err
	}
	return request, nil
}

func (r *FileRepository) Get(_ context.Context, id string) (Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	requests, err := r.load()
	if err != nil {
		return Request{}, err
	}
	request, ok := requests[id]
	if !ok {
		return Request{}, ErrNotFound
	}
	return request, nil
}

func (r *FileRepository) ListPending(_ context.Context) ([]Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	requests, err := r.load()
	if err != nil {
		return nil, err
	}
	results := make([]Request, 0, len(requests))
	for _, request := range requests {
		if request.Status == StatusPending {
			results = append(results, request)
		}
	}
	return results, nil
}

func (r *FileRepository) Update(_ context.Context, request Request) (Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	requests, err := r.load()
	if err != nil {
		return Request{}, err
	}
	existing, ok := requests[request.ID]
	if !ok {
		return Request{}, ErrNotFound
	}
	if existing.Status != StatusPending && existing.Status != request.Status {
		return Request{}, ErrConflict
	}
	requests[request.ID] = request
	if err := r.save(requests); err != nil {
		return Request{}, err
	}
	return request, nil
}

func (r *FileRepository) load() (map[string]Request, error) {
	if r.path == "" {
		return nil, fmt.Errorf("%w: file path is required", ErrInvalidInput)
	}
	raw, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]Request), nil
		}
		return nil, fmt.Errorf("read approval store %q: %w", r.path, err)
	}
	if len(raw) == 0 {
		return make(map[string]Request), nil
	}
	var requests map[string]Request
	if err := json.Unmarshal(raw, &requests); err != nil {
		return nil, fmt.Errorf("decode approval store %q: %w", r.path, err)
	}
	if requests == nil {
		requests = make(map[string]Request)
	}
	return requests, nil
}

func (r *FileRepository) save(requests map[string]Request) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return fmt.Errorf("create approval store directory: %w", err)
	}
	raw, err := json.MarshalIndent(requests, "", "  ")
	if err != nil {
		return fmt.Errorf("encode approval store: %w", err)
	}
	if err := os.WriteFile(r.path, raw, 0o600); err != nil {
		return fmt.Errorf("write approval store %q: %w", r.path, err)
	}
	return nil
}