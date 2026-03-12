package approval

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusDenied   Status = "denied"
)

var (
	ErrNotFound     = errors.New("approval request not found")
	ErrConflict     = errors.New("approval request already resolved differently")
	ErrInvalidInput = errors.New("invalid approval input")
)

// Request is one approval workflow record.
type Request struct {
	ID         string         `json:"id"`
	Status     Status         `json:"status"`
	Server     string         `json:"server"`
	Tool       string         `json:"tool"`
	Method     string         `json:"method"`
	Reason     string         `json:"reason"`
	RuleName   string         `json:"rule_name,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	Arguments  map[string]any `json:"arguments,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	CreatedBy  string         `json:"created_by,omitempty"`
	ResolvedAt *time.Time     `json:"resolved_at,omitempty"`
	ResolvedBy string         `json:"resolved_by,omitempty"`
	Resolution string         `json:"resolution,omitempty"`
}

// CreateInput is the input required to create a pending approval request.
type CreateInput struct {
	Server    string
	Tool      string
	Method    string
	Reason    string
	RuleName  string
	RequestID string
	Arguments map[string]any
	Actor     string
}

// ResolveInput records an approval or denial decision.
type ResolveInput struct {
	ID     string
	Actor  string
	Reason string
}

// Repository persists approval requests.
type Repository interface {
	Create(ctx context.Context, request Request) (Request, error)
	Get(ctx context.Context, id string) (Request, error)
	ListPending(ctx context.Context) ([]Request, error)
	Update(ctx context.Context, request Request) (Request, error)
}

// Service coordinates approval creation and resolution.
type Service struct {
	repo Repository
	now  func() time.Time
	ids  func() string
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }, ids: newID}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Request, error) {
	if err := validateCreateInput(input); err != nil {
		return Request{}, err
	}
	request := Request{
		ID:        s.ids(),
		Status:    StatusPending,
		Server:    input.Server,
		Tool:      input.Tool,
		Method:    input.Method,
		Reason:    input.Reason,
		RuleName:  input.RuleName,
		RequestID: input.RequestID,
		Arguments: cloneMap(input.Arguments),
		CreatedAt: s.now(),
		CreatedBy: input.Actor,
	}
	return s.repo.Create(ctx, request)
}

func (s *Service) ListPending(ctx context.Context) ([]Request, error) {
	requests, err := s.repo.ListPending(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(requests, func(i, j int) bool {
		if requests[i].CreatedAt.Equal(requests[j].CreatedAt) {
			return requests[i].ID < requests[j].ID
		}
		return requests[i].CreatedAt.Before(requests[j].CreatedAt)
	})
	return requests, nil
}

func (s *Service) Approve(ctx context.Context, input ResolveInput) (Request, error) {
	return s.resolve(ctx, input, StatusApproved)
}

func (s *Service) Deny(ctx context.Context, input ResolveInput) (Request, error) {
	return s.resolve(ctx, input, StatusDenied)
}

func (s *Service) resolve(ctx context.Context, input ResolveInput, status Status) (Request, error) {
	if strings.TrimSpace(input.ID) == "" {
		return Request{}, fmt.Errorf("%w: id is required", ErrInvalidInput)
	}
	request, err := s.repo.Get(ctx, input.ID)
	if err != nil {
		return Request{}, err
	}
	if request.Status == status {
		return request, nil
	}
	if request.Status != StatusPending {
		return Request{}, ErrConflict
	}
	now := s.now()
	request.Status = status
	request.ResolvedAt = &now
	request.ResolvedBy = input.Actor
	request.Resolution = input.Reason
	return s.repo.Update(ctx, request)
}

func validateCreateInput(input CreateInput) error {
	if strings.TrimSpace(input.Server) == "" {
		return fmt.Errorf("%w: server is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Tool) == "" {
		return fmt.Errorf("%w: tool is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Method) == "" {
		return fmt.Errorf("%w: method is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
	}
	return nil
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	output := make(map[string]any, len(input))
	for key, value := range input {
		switch v := value.(type) {
		case map[string]any:
			output[key] = cloneMap(v)
		case []any:
			items := make([]any, len(v))
			for i := range v {
				items[i] = v[i]
			}
			output[key] = items
		default:
			output[key] = value
		}
	}
	return output
}

// MemoryRepository keeps approval requests in memory for tests and embedding.
type MemoryRepository struct {
	mu       sync.Mutex
	requests map[string]Request
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{requests: make(map[string]Request)}
}

func (r *MemoryRepository) Create(_ context.Context, request Request) (Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[request.ID] = request
	return request, nil
}

func (r *MemoryRepository) Get(_ context.Context, id string) (Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	request, ok := r.requests[id]
	if !ok {
		return Request{}, ErrNotFound
	}
	return request, nil
}

func (r *MemoryRepository) ListPending(_ context.Context) ([]Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	results := make([]Request, 0, len(r.requests))
	for _, request := range r.requests {
		if request.Status == StatusPending {
			results = append(results, request)
		}
	}
	return results, nil
}

func (r *MemoryRepository) Update(_ context.Context, request Request) (Request, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.requests[request.ID]
	if !ok {
		return Request{}, ErrNotFound
	}
	if existing.Status != StatusPending && existing.Status != request.Status {
		return Request{}, ErrConflict
	}
	r.requests[request.ID] = request
	return request, nil
}