package database

import (
	"context"
	"database/sql"
	"fmt"
)

func migrateTables(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			role TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			quota_bytes INTEGER NOT NULL,
			share_ttl_days INTEGER NOT NULL DEFAULT 30,
			storage_used_bytes INTEGER NOT NULL DEFAULT 0,
			disabled_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
		CREATE INDEX IF NOT EXISTS idx_users_disabled_at ON users(disabled_at);

		CREATE TABLE IF NOT EXISTS media_files (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL DEFAULT '',
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
			deleted_at TEXT,
			deleted_reason TEXT NOT NULL DEFAULT ''
		);
		CREATE INDEX IF NOT EXISTS idx_media_deleted_at ON media_files(deleted_at);

		CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			key_prefix TEXT NOT NULL,
			scopes TEXT NOT NULL,
			created_at TEXT NOT NULL,
			last_used_at TEXT,
			revoked_at TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);

		CREATE TABLE IF NOT EXISTS transfers (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			title TEXT NOT NULL DEFAULT '',
			message TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			revoked_at TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_transfers_owner ON transfers(owner_id, revoked_at, created_at);

		CREATE TABLE IF NOT EXISTS transfer_files (
			transfer_id TEXT NOT NULL,
			file_id TEXT NOT NULL,
			created_at TEXT NOT NULL,
			PRIMARY KEY (transfer_id, file_id)
		);
		CREATE INDEX IF NOT EXISTS idx_transfer_files_file ON transfer_files(file_id);

		CREATE TABLE IF NOT EXISTS shares (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			code TEXT NOT NULL UNIQUE,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			never_expires INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			revoked_at TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_shares_owner ON shares(owner_id, revoked_at, created_at);
		CREATE INDEX IF NOT EXISTS idx_shares_target ON shares(target_type, target_id, revoked_at);
	`)
	if err != nil {
		return fmt.Errorf("migrar sqlite: %w", err)
	}
	return nil
}

func createCompatibilityIndexes(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_media_owner_active ON media_files(owner_id, deleted_at, created_at);
		CREATE INDEX IF NOT EXISTS idx_api_keys_owner ON api_keys(owner_id, revoked_at);
	`)
	if err != nil {
		return fmt.Errorf("crear indices compatibles: %w", err)
	}
	return nil
}

func ensureColumn(ctx context.Context, db *sql.DB, table string, column string, definition string) error {
	exists, err := columnExists(ctx, db, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)); err != nil {
		return fmt.Errorf("agregar columna %s.%s: %w", table, column, err)
	}
	return nil
}

func columnExists(ctx context.Context, db *sql.DB, table string, column string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, fmt.Errorf("leer columnas %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		matches, err := columnMatches(rows, table, column)
		if err != nil {
			return false, err
		}
		if matches {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterar columnas %s: %w", table, err)
	}
	return false, nil
}

func columnMatches(rows *sql.Rows, table string, column string) (bool, error) {
	var cid int
	var name string
	var dataType string
	var notNull int
	var defaultValue sql.NullString
	var pk int
	if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
		return false, fmt.Errorf("leer columna %s: %w", table, err)
	}
	return name == column, nil
}
