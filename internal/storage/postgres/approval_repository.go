package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/agentfence/agentfence/internal/approval"
)

type ApprovalRepository struct { db *sql.DB }

func NewApprovalRepository(db *sql.DB) *ApprovalRepository { return &ApprovalRepository{db: db} }

func (r *ApprovalRepository) Create(ctx context.Context, request approval.Request) (approval.Request, error) {
	arguments, err := json.Marshal(request.Arguments)
	if err != nil {
		return approval.Request{}, fmt.Errorf("marshal approval arguments: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO approval_requests (
			id, status, server_name, tool_name, method_name, reason, rule_name,
			request_id, arguments, created_at, created_by, resolved_at, resolved_by, resolution
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, request.ID, request.Status, request.Server, request.Tool, request.Method, request.Reason, nullString(request.RuleName), nullString(request.RequestID), arguments, request.CreatedAt, nullString(request.CreatedBy), request.ResolvedAt, nullString(request.ResolvedBy), nullString(request.Resolution))
	if err != nil {
		return approval.Request{}, fmt.Errorf("insert approval request: %w", err)
	}
	return request, nil
}

func (r *ApprovalRepository) Get(ctx context.Context, id string) (approval.Request, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, status, server_name, tool_name, method_name, reason, rule_name, request_id,
		arguments, created_at, created_by, resolved_at, resolved_by, resolution
		FROM approval_requests WHERE id = $1
	`, id)
	request, err := scanApproval(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return approval.Request{}, approval.ErrNotFound
		}
		return approval.Request{}, err
	}
	return request, nil
}

func (r *ApprovalRepository) ListPending(ctx context.Context) ([]approval.Request, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, status, server_name, tool_name, method_name, reason, rule_name, request_id,
		arguments, created_at, created_by, resolved_at, resolved_by, resolution
		FROM approval_requests WHERE status = $1 ORDER BY created_at ASC, id ASC
	`, approval.StatusPending)
	if err != nil {
		return nil, fmt.Errorf("query pending approvals: %w", err)
	}
	defer rows.Close()
	requests := make([]approval.Request, 0)
	for rows.Next() {
		request, err := scanApproval(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	sort.Slice(requests, func(i, j int) bool {
		if requests[i].CreatedAt.Equal(requests[j].CreatedAt) {
			return requests[i].ID < requests[j].ID
		}
		return requests[i].CreatedAt.Before(requests[j].CreatedAt)
	})
	return requests, rows.Err()
}

func (r *ApprovalRepository) Update(ctx context.Context, request approval.Request) (approval.Request, error) {
	arguments, err := json.Marshal(request.Arguments)
	if err != nil {
		return approval.Request{}, fmt.Errorf("marshal approval arguments: %w", err)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE approval_requests
		SET status = $2, reason = $3, rule_name = $4, request_id = $5, arguments = $6,
			created_at = $7, created_by = $8, resolved_at = $9, resolved_by = $10, resolution = $11
		WHERE id = $1
	`, request.ID, request.Status, request.Reason, nullString(request.RuleName), nullString(request.RequestID), arguments, request.CreatedAt, nullString(request.CreatedBy), request.ResolvedAt, nullString(request.ResolvedBy), nullString(request.Resolution))
	if err != nil {
		return approval.Request{}, fmt.Errorf("update approval request: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return approval.Request{}, fmt.Errorf("approval rows affected: %w", err)
	}
	if affected == 0 {
		return approval.Request{}, approval.ErrNotFound
	}
	return request, nil
}

type approvalScanner interface{ Scan(dest ...any) error }

func scanApproval(scanner approvalScanner) (approval.Request, error) {
	var request approval.Request
	var status string
	var ruleName sql.NullString
	var requestID sql.NullString
	var arguments []byte
	var createdBy sql.NullString
	var resolvedAt sql.NullTime
	var resolvedBy sql.NullString
	var resolution sql.NullString
	if err := scanner.Scan(&request.ID, &status, &request.Server, &request.Tool, &request.Method, &request.Reason, &ruleName, &requestID, &arguments, &request.CreatedAt, &createdBy, &resolvedAt, &resolvedBy, &resolution); err != nil {
		return approval.Request{}, err
	}
	request.Status = approval.Status(status)
	request.RuleName = ruleName.String
	request.RequestID = requestID.String
	request.CreatedBy = createdBy.String
	if resolvedAt.Valid {
		resolved := resolvedAt.Time
		request.ResolvedAt = &resolved
	}
	request.ResolvedBy = resolvedBy.String
	request.Resolution = resolution.String
	if len(arguments) > 0 {
		if err := json.Unmarshal(arguments, &request.Arguments); err != nil {
			return approval.Request{}, fmt.Errorf("decode approval arguments: %w", err)
		}
	}
	return request, nil
}
