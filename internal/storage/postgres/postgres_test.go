package postgres

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/agentfence/agentfence/internal/audit"
	"github.com/agentfence/agentfence/internal/approval"
)

func TestSplitMigration(t *testing.T) {
	up, down := splitMigration("-- +agentfence Up\nCREATE TABLE x ();\n-- +agentfence Down\nDROP TABLE x;")
	if up != "CREATE TABLE x ();" || down != "DROP TABLE x;" {
		t.Fatalf("splitMigration() = (%q, %q), want up/down SQL", up, down)
	}
}

func TestLoadMigrations(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(migrations) == 0 || migrations[0].version != "001_init" {
		t.Fatalf("migrations = %+v, want 001_init present", migrations)
	}
}

func TestDecodeAuditEvent(t *testing.T) {
	input := audit.Event{Kind: audit.EventKindUpstreamCall, Request: audit.RequestContext{ID: "req-1", Server: "deployer", Tool: "deploy", Method: "tools/call"}}
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	decoded, err := decodeAuditEvent(raw)
	if err != nil {
		t.Fatalf("decodeAuditEvent() error = %v", err)
	}
	if !reflect.DeepEqual(decoded.Request, input.Request) || decoded.Kind != input.Kind {
		t.Fatalf("decoded = %+v, want %+v", decoded, input)
	}
}

func TestScanApproval(t *testing.T) {
	createdAt := time.Date(2026, time.March, 12, 13, 0, 0, 0, time.UTC)
	resolvedAt := time.Date(2026, time.March, 12, 14, 0, 0, 0, time.UTC)
	arguments, _ := json.Marshal(map[string]any{"environment": "prod"})
	scanner := fakeScanner{values: []any{
		"apr_1", "approved", "deployer", "deploy", "tools/call", "needs approval",
		sql.NullString{String: "approval-prod", Valid: true}, sql.NullString{String: "req-1", Valid: true}, arguments,
		createdAt, sql.NullString{String: "gateway", Valid: true}, sql.NullTime{Time: resolvedAt, Valid: true}, sql.NullString{String: "alice", Valid: true}, sql.NullString{String: "ok", Valid: true},
	}}
	request, err := scanApproval(scanner)
	if err != nil {
		t.Fatalf("scanApproval() error = %v", err)
	}
	if request.ID != "apr_1" || request.Status != approval.StatusApproved || request.Arguments["environment"] != "prod" {
		t.Fatalf("request = %+v, want decoded approval", request)
	}
	if request.ResolvedAt == nil || !request.ResolvedAt.Equal(resolvedAt) {
		t.Fatalf("ResolvedAt = %v, want %v", request.ResolvedAt, resolvedAt)
	}
}

type fakeScanner struct{ values []any }

func (s fakeScanner) Scan(dest ...any) error {
	for i := range dest {
		switch target := dest[i].(type) {
		case *string:
			*target = s.values[i].(string)
		case *sql.NullString:
			*target = s.values[i].(sql.NullString)
		case *[]byte:
			*target = s.values[i].([]byte)
		case *time.Time:
			*target = s.values[i].(time.Time)
		case *sql.NullTime:
			*target = s.values[i].(sql.NullTime)
		default:
			return sql.ErrConnDone
		}
	}
	return nil
}
