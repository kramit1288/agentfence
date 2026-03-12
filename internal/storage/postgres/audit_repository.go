package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/agentfence/agentfence/internal/audit"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository { return &AuditRepository{db: db} }

func (r *AuditRepository) Record(ctx context.Context, event audit.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	var latencyMS sql.NullInt64
	if event.Upstream.Latency > 0 {
		latencyMS = sql.NullInt64{Int64: event.Upstream.Latency.Milliseconds(), Valid: true}
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO audit_events (
			event_time, kind, request_id, server_name, tool_name, method_name,
			decision_action, decision_reason, decision_rule_name, decision_allowed,
			upstream_target, upstream_outcome, upstream_status_code, upstream_latency_ms,
			upstream_error, upstream_forwarded, payload
		) VALUES (
			$1,$2,$3,$4,$5,$6,
			$7,$8,$9,$10,
			$11,$12,$13,$14,
			$15,$16,$17
		)
	`, event.Timestamp, event.Kind, nullString(event.Request.ID), event.Request.Server, event.Request.Tool, event.Request.Method,
		nullString(string(event.Decision.Action)), nullString(event.Decision.Reason), nullString(event.Decision.RuleName), nullBool(event.Decision.Action != "", event.Decision.Allowed),
		nullString(event.Upstream.Target), nullString(event.Upstream.Outcome), nullInt(event.Upstream.HTTPStatusCode), latencyMS,
		nullString(event.Upstream.Error), nullBool(event.Upstream.Outcome != "", event.Upstream.Forwarded), payload)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func (r *AuditRepository) ListRecent(ctx context.Context, limit int) ([]audit.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `SELECT payload FROM audit_events ORDER BY event_time DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent audit events: %w", err)
	}
	defer rows.Close()
	events := make([]audit.Event, 0, limit)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan audit event: %w", err)
		}
		event, err := decodeAuditEvent(payload)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func decodeAuditEvent(payload []byte) (audit.Event, error) {
	var event audit.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return audit.Event{}, fmt.Errorf("decode audit event payload: %w", err)
	}
	return event, nil
}

func nullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

func nullInt(value int) sql.NullInt64 {
	if value == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(value), Valid: true}
}

func nullBool(valid bool, value bool) sql.NullBool {
	if !valid {
		return sql.NullBool{}
	}
	return sql.NullBool{Bool: value, Valid: true}
}
