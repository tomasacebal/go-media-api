package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestMigrateAddsOwnerColumnsBeforeIndexes(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "metadata.sqlite"))
	if err != nil {
		t.Fatalf("abrir sqlite fallo: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE media_files (
			id TEXT PRIMARY KEY,
			original_name TEXT NOT NULL,
			stored_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			extension TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			storage_driver TEXT NOT NULL,
			storage_path TEXT NOT NULL UNIQUE,
			public_url TEXT NOT NULL,
			visibility TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			category TEXT NOT NULL DEFAULT '',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			deleted_at TEXT
		);
		CREATE TABLE api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			key_prefix TEXT NOT NULL,
			scopes TEXT NOT NULL,
			created_at TEXT NOT NULL,
			last_used_at TEXT,
			revoked_at TEXT
		);
	`); err != nil {
		t.Fatalf("crear esquema viejo fallo: %v", err)
	}

	if err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate fallo: %v", err)
	}
	if exists, err := columnExists(context.Background(), db, "media_files", "owner_id"); err != nil || !exists {
		t.Fatalf("owner_id media esperado, exists=%v err=%v", exists, err)
	}
	if exists, err := columnExists(context.Background(), db, "api_keys", "owner_id"); err != nil || !exists {
		t.Fatalf("owner_id api keys esperado, exists=%v err=%v", exists, err)
	}
	if exists, err := columnExists(context.Background(), db, "users", "share_ttl_days"); err != nil || !exists {
		t.Fatalf("share_ttl_days esperado, exists=%v err=%v", exists, err)
	}
	if exists, err := columnExists(context.Background(), db, "shares", "never_expires"); err != nil || !exists {
		t.Fatalf("never_expires esperado, exists=%v err=%v", exists, err)
	}
}
