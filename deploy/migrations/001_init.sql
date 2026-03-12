-- +agentfence Up
CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_events (
  id BIGSERIAL PRIMARY KEY,
  event_time TIMESTAMPTZ NOT NULL,
  kind TEXT NOT NULL,
  request_id TEXT,
  server_name TEXT NOT NULL,
  tool_name TEXT NOT NULL,
  method_name TEXT NOT NULL,
  decision_action TEXT,
  decision_reason TEXT,
  decision_rule_name TEXT,
  decision_allowed BOOLEAN,
  upstream_target TEXT,
  upstream_outcome TEXT,
  upstream_status_code INTEGER,
  upstream_latency_ms BIGINT,
  upstream_error TEXT,
  upstream_forwarded BOOLEAN,
  payload JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS audit_events_event_time_idx ON audit_events (event_time DESC);
CREATE INDEX IF NOT EXISTS audit_events_server_name_idx ON audit_events (server_name);
CREATE INDEX IF NOT EXISTS audit_events_kind_idx ON audit_events (kind);

CREATE TABLE IF NOT EXISTS approval_requests (
  id TEXT PRIMARY KEY,
  status TEXT NOT NULL,
  server_name TEXT NOT NULL,
  tool_name TEXT NOT NULL,
  method_name TEXT NOT NULL,
  reason TEXT NOT NULL,
  rule_name TEXT,
  request_id TEXT,
  arguments JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  created_by TEXT,
  resolved_at TIMESTAMPTZ,
  resolved_by TEXT,
  resolution TEXT
);

CREATE INDEX IF NOT EXISTS approval_requests_status_idx ON approval_requests (status, created_at ASC);
CREATE INDEX IF NOT EXISTS approval_requests_created_at_idx ON approval_requests (created_at DESC);
-- +agentfence Down
DROP TABLE IF EXISTS approval_requests;
DROP TABLE IF EXISTS audit_events;
DROP TABLE IF EXISTS schema_migrations;
