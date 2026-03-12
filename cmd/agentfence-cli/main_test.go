package main

import (
	"strings"
	"testing"
)

func TestResolveApprovalRequiresActor(t *testing.T) {
	err := resolveApproval([]string{"--store", "testdata/approvals.json", "apr_123"}, true)
	if err == nil {
		t.Fatal("resolveApproval() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "actor is required") {
		t.Fatalf("resolveApproval() error = %v, want actor required", err)
	}
}