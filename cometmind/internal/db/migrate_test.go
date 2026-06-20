package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEnsureSchemaUpgradesContextSummaryColumns(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "migrate-v9.db")
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.ExecContext(ctx, "PRAGMA user_version = 9"); err != nil {
		t.Fatalf("set user_version: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `CREATE TABLE sessions (
		id TEXT PRIMARY KEY,
		workspace_id TEXT NOT NULL,
		title TEXT NOT NULL DEFAULT '',
		model_id TEXT NOT NULL,
		provider_id TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		token_usage TEXT NOT NULL DEFAULT '{}',
		parent_session_id TEXT,
		purpose TEXT NOT NULL DEFAULT '',
		delegation_status TEXT NOT NULL DEFAULT '',
		output_summary TEXT NOT NULL DEFAULT '',
		acp_session_id TEXT NOT NULL DEFAULT '',
		pending_question TEXT NOT NULL DEFAULT '',
		pinned INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL DEFAULT 0,
		updated_at INTEGER NOT NULL DEFAULT 0
	)`); err != nil {
		t.Fatalf("create sessions table: %v", err)
	}

	if err := EnsureSchema(ctx, conn); err != nil {
		t.Fatalf("EnsureSchema() error = %v", err)
	}

	var version int
	if err := conn.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != schemaVersion {
		t.Fatalf("user_version = %d, want %d", version, schemaVersion)
	}

	columns := map[string]bool{}
	rows, err := conn.QueryContext(ctx, "PRAGMA table_info(sessions)")
	if err != nil {
		t.Fatalf("table_info: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		columns[name] = true
	}
	for _, col := range []string{"context_summary", "compacted_until_message_id", "context_summary_updated_at"} {
		if !columns[col] {
			t.Fatalf("sessions missing column %q after migration", col)
		}
	}
}
