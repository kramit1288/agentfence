package approval

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateApprovalRequest(t *testing.T) {
	svc := testService(NewMemoryRepository())

	request, err := svc.Create(context.Background(), CreateInput{
		Server:    "deployer",
		Tool:      "deploy",
		Method:    "tools/call",
		Reason:    "production deploy needs approval",
		RuleName:  "approval-prod",
		RequestID: "req-1",
		Arguments: map[string]any{"environment": "prod"},
		Actor:     "gateway",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if request.ID == "" || request.Status != StatusPending {
		t.Fatalf("request = %+v, want pending request with id", request)
	}
	if request.CreatedBy != "gateway" {
		t.Fatalf("CreatedBy = %q, want gateway", request.CreatedBy)
	}
}

func TestApproveApprovalRequest(t *testing.T) {
	svc := testService(NewMemoryRepository())
	request := mustCreate(t, svc)

	approved, err := svc.Approve(context.Background(), ResolveInput{ID: request.ID, Actor: "alice", Reason: "looks safe"})
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if approved.Status != StatusApproved || approved.ResolvedBy != "alice" {
		t.Fatalf("approved = %+v, want approved by alice", approved)
	}
	if approved.ResolvedAt == nil {
		t.Fatal("ResolvedAt = nil, want timestamp")
	}
}

func TestDenyApprovalRequest(t *testing.T) {
	svc := testService(NewMemoryRepository())
	request := mustCreate(t, svc)

	denied, err := svc.Deny(context.Background(), ResolveInput{ID: request.ID, Actor: "bob", Reason: "too risky"})
	if err != nil {
		t.Fatalf("Deny() error = %v", err)
	}
	if denied.Status != StatusDenied || denied.ResolvedBy != "bob" {
		t.Fatalf("denied = %+v, want denied by bob", denied)
	}
}

func TestApproveIsIdempotent(t *testing.T) {
	svc := testService(NewMemoryRepository())
	request := mustCreate(t, svc)

	first, err := svc.Approve(context.Background(), ResolveInput{ID: request.ID, Actor: "alice", Reason: "ok"})
	if err != nil {
		t.Fatalf("Approve() first error = %v", err)
	}
	second, err := svc.Approve(context.Background(), ResolveInput{ID: request.ID, Actor: "charlie", Reason: "duplicate"})
	if err != nil {
		t.Fatalf("Approve() second error = %v", err)
	}
	if second.Status != StatusApproved || second.ResolvedBy != first.ResolvedBy {
		t.Fatalf("second = %+v, want unchanged approved record", second)
	}
}

func TestInvalidApprovalID(t *testing.T) {
	svc := testService(NewMemoryRepository())

	_, err := svc.Approve(context.Background(), ResolveInput{ID: "missing"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Approve() error = %v, want ErrNotFound", err)
	}
}

func TestConflictingResolutionFails(t *testing.T) {
	svc := testService(NewMemoryRepository())
	request := mustCreate(t, svc)
	if _, err := svc.Approve(context.Background(), ResolveInput{ID: request.ID, Actor: "alice"}); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}

	_, err := svc.Deny(context.Background(), ResolveInput{ID: request.ID, Actor: "bob"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("Deny() error = %v, want ErrConflict", err)
	}
}

func TestMemoryRepositoryRejectsStaleConflictingUpdate(t *testing.T) {
	repo := NewMemoryRepository()
	request := Request{ID: "apr_1", Status: StatusApproved, Server: "deployer", Tool: "deploy", Method: "tools/call", Reason: "needs approval"}
	if _, err := repo.Create(context.Background(), request); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	request.Status = StatusDenied
	if _, err := repo.Update(context.Background(), request); !errors.Is(err, ErrConflict) {
		t.Fatalf("Update() error = %v, want ErrConflict", err)
	}
}

func TestFileRepositoryPersistsPendingRequests(t *testing.T) {
	path := filepath.Join(t.TempDir(), "approvals.json")
	repo := NewFileRepository(path)
	svc := testService(repo)

	request, err := svc.Create(context.Background(), CreateInput{
		Server: "deployer",
		Tool:   "deploy",
		Method: "tools/call",
		Reason: "approval required",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	other := NewService(NewFileRepository(path))
	pending, err := other.ListPending(context.Background())
	if err != nil {
		t.Fatalf("ListPending() error = %v", err)
	}
	if len(pending) != 1 || pending[0].ID != request.ID {
		t.Fatalf("pending = %+v, want created request", pending)
	}
}

func TestFileRepositoryRejectsStaleConflictingUpdate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "approvals.json")
	repo := NewFileRepository(path)
	request := Request{ID: "apr_1", Status: StatusApproved, Server: "deployer", Tool: "deploy", Method: "tools/call", Reason: "needs approval"}
	if _, err := repo.Create(context.Background(), request); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	request.Status = StatusDenied
	if _, err := repo.Update(context.Background(), request); !errors.Is(err, ErrConflict) {
		t.Fatalf("Update() error = %v, want ErrConflict", err)
	}
}

func testService(repo Repository) *Service {
	svc := NewService(repo)
	fixed := time.Date(2026, time.March, 12, 13, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return fixed }
	svc.ids = func() string { return "apr_test_1" }
	return svc
}

func mustCreate(t *testing.T, svc *Service) Request {
	t.Helper()
	request, err := svc.Create(context.Background(), CreateInput{
		Server:    "deployer",
		Tool:      "deploy",
		Method:    "tools/call",
		Reason:    "production deploy needs approval",
		RequestID: "req-1",
		Arguments: map[string]any{"environment": "prod"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	return request
}