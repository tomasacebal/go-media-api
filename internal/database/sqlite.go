package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// OpenSQLite abre SQLite, configura la conexion y ejecuta migraciones.
//
// Args:
//   - dbPath: path del archivo SQLite.
//
// Returns:
//   - Conexion abierta, funcion de cierre o error.
func OpenSQLite(dbPath string) (*sql.DB, func() error, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, nil, fmt.Errorf("crear carpeta sqlite: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("abrir sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;"); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("configurar sqlite: %w", err)
	}
	if err := Migrate(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	return db, db.Close, nil
}

// Migrate crea y actualiza las tablas del producto.
//
// Args:
//   - ctx: contexto de la operacion.
//   - db: conexion SQLite abierta.
//
// Returns:
//   - Error si una migracion falla.
func Migrate(ctx context.Context, db *sql.DB) error {
	if err := migrateTables(ctx, db); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "media_files", "owner_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "media_files", "deleted_reason", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "api_keys", "owner_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "users", "share_ttl_days", "INTEGER NOT NULL DEFAULT 30"); err != nil {
		return err
	}
	if err := ensureColumn(ctx, db, "shares", "never_expires", "INTEGER NOT NULL DEFAULT 0"); err != nil {
		return err
	}
	if err := createCompatibilityIndexes(ctx, db); err != nil {
		return err
	}
	return nil
}

// AssignLegacyOwnership asigna datos viejos al admin bootstrap.
//
// Args:
//   - ctx: contexto de la operacion.
//   - db: conexion SQLite abierta.
//   - ownerID: usuario dueño de datos sin owner.
//
// Returns:
//   - Error si no pudo actualizar datos heredados.
func AssignLegacyOwnership(ctx context.Context, db *sql.DB, ownerID string) error {
	if _, err := db.ExecContext(ctx, "UPDATE media_files SET owner_id = ? WHERE owner_id = ''", ownerID); err != nil {
		return fmt.Errorf("asignar media heredada: %w", err)
	}
	if _, err := db.ExecContext(ctx, "UPDATE api_keys SET owner_id = ? WHERE owner_id = ''", ownerID); err != nil {
		return fmt.Errorf("asignar api keys heredadas: %w", err)
	}
	return RecalculateStorageUsage(ctx, db)
}

// RecalculateStorageUsage recalcula uso de storage desde archivos activos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - db: conexion SQLite abierta.
//
// Returns:
//   - Error si no pudo sincronizar usos.
func RecalculateStorageUsage(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, "UPDATE users SET storage_used_bytes = 0"); err != nil {
		return fmt.Errorf("resetear uso storage: %w", err)
	}
	_, err := db.ExecContext(ctx, `
		UPDATE users
		SET storage_used_bytes = COALESCE((
			SELECT SUM(size_bytes)
			FROM media_files
			WHERE media_files.owner_id = users.id AND media_files.deleted_at IS NULL
		), 0)
	`)
	if err != nil {
		return fmt.Errorf("recalcular uso storage: %w", err)
	}
	return nil
}
