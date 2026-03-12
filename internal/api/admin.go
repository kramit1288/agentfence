package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/agentfence/agentfence/internal/approval"
	"github.com/agentfence/agentfence/internal/audit"
)

type AuditReader interface {
	ListRecent(ctx context.Context, limit int) ([]audit.Event, error)
}

type ApprovalReader interface {
	ListPending(ctx context.Context) ([]approval.Request, error)
}

type PolicyStatus struct {
	Configured bool   `json:"configured"`
	RuleCount  int    `json:"ruleCount"`
	Source     string `json:"source"`
	Summary    string `json:"summary"`
}

type PolicyStatusProvider interface {
	Status(ctx context.Context) (PolicyStatus, error)
}

type AdminDeps struct {
	AuditReader          AuditReader
	ApprovalReader       ApprovalReader
	PolicyStatusProvider PolicyStatusProvider
}

func registerAdminRoutes(mux *http.ServeMux, deps AdminDeps) {
	mux.HandleFunc("/api/admin/audit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, http.MethodGet)
			return
		}
		if deps.AuditReader == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": []audit.Event{}, "available": false})
			return
		}
		limit := 50
		if raw := r.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 200 {
				limit = parsed
			}
		}
		items, err := deps.AuditReader.ListRecent(r.Context(), limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "available": true})
	})

	mux.HandleFunc("/api/admin/approvals", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, http.MethodGet)
			return
		}
		if deps.ApprovalReader == nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": []approval.Request{}, "available": false})
			return
		}
		items, err := deps.ApprovalReader.ListPending(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "available": true})
	})

	mux.HandleFunc("/api/admin/policy", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, http.MethodGet)
			return
		}
		if deps.PolicyStatusProvider == nil {
			writeJSON(w, http.StatusOK, PolicyStatus{Configured: false, Source: "unconfigured", Summary: "No runtime policy status provider configured."})
			return
		}
		status, err := deps.PolicyStatusProvider.Status(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, status)
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
