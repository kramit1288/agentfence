package audit

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestFileRepositoryRecordAndListRecent(t *testing.T) {
	repo := NewFileRepository(filepath.Join(t.TempDir(), "audit.json"))
	early := Event{Timestamp: time.Date(2026, 3, 10, 9, 0, 0, 0, time.UTC), Kind: EventKindPolicyDecision, Request: RequestContext{Server: "github-demo", Tool: "repos/get", Method: "tools/call"}}
	late := Event{Timestamp: time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC), Kind: EventKindUpstreamCall, Request: RequestContext{Server: "github-demo", Tool: "pulls/list", Method: "tools/call"}}

	if err := repo.Record(context.Background(), early); err != nil {
		t.Fatalf("Record(early) error = %v", err)
	}
	if err := repo.Record(context.Background(), late); err != nil {
		t.Fatalf("Record(late) error = %v", err)
	}

	events, err := repo.ListRecent(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListRecent() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].Kind != EventKindUpstreamCall {
		t.Fatalf("events[0].Kind = %q, want %q", events[0].Kind, EventKindUpstreamCall)
	}
}

func TestFileRepositoryListRecentLimit(t *testing.T) {
	repo := NewFileRepository(filepath.Join(t.TempDir(), "audit.json"))
	for i := 0; i < 3; i++ {
		event := Event{Timestamp: time.Date(2026, 3, 10, 9+i, 0, 0, 0, time.UTC), Kind: EventKindPolicyDecision, Request: RequestContext{Server: "github-demo", Tool: "repos/get", Method: "tools/call"}}
		if err := repo.Record(context.Background(), event); err != nil {
			t.Fatalf("Record(%d) error = %v", i, err)
		}
	}

	events, err := repo.ListRecent(context.Background(), 2)
	if err != nil {
		t.Fatalf("ListRecent() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if !events[0].Timestamp.After(events[1].Timestamp) {
		t.Fatalf("timestamps = %v, want descending order", events)
	}
}