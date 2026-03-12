package postgres

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type DB struct {
	SQL *sql.DB
}

func Open(ctx context.Context, dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &DB{SQL: db}, nil
}

func (db *DB) Close() error { return db.SQL.Close() }

func (db *DB) Migrate(ctx context.Context) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}
	if _, err := db.SQL.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	applied, err := appliedVersions(ctx, db.SQL)
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		if _, ok := applied[migration.version]; ok {
			continue
		}
		if strings.TrimSpace(migration.upSQL) == "" {
			continue
		}
		tx, err := db.SQL.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", migration.version, err)
		}
		if _, err := tx.ExecContext(ctx, migration.upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", migration.version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, migration.version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", migration.version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", migration.version, err)
		}
	}
	return nil
}

type migration struct {
	version string
	upSQL   string
}

func loadMigrations() ([]migration, error) {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}
	migrations := make([]migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		raw, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		upSQL, _ := splitMigration(string(raw))
		migrations = append(migrations, migration{version: strings.TrimSuffix(entry.Name(), ".sql"), upSQL: upSQL})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	return migrations, nil
}

func splitMigration(raw string) (string, string) {
	upMarker := "-- +agentfence Up"
	downMarker := "-- +agentfence Down"
	upIndex := strings.Index(raw, upMarker)
	downIndex := strings.Index(raw, downMarker)
	if upIndex == -1 {
		return strings.TrimSpace(raw), ""
	}
	start := upIndex + len(upMarker)
	if downIndex == -1 || downIndex < start {
		return strings.TrimSpace(raw[start:]), ""
	}
	return strings.TrimSpace(raw[start:downIndex]), strings.TrimSpace(raw[downIndex+len(downMarker):])
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("query schema migrations: %w", err)
	}
	defer rows.Close()
	versions := make(map[string]struct{})
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan schema migration: %w", err)
		}
		versions[version] = struct{}{}
	}
	return versions, rows.Err()
}
