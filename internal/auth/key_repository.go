package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SQLiteKeyRepository persiste api keys en SQLite.
//
// Args:
//   - db: conexion SQL compartida.
//
// Returns:
//   - Repository listo para api keys.
type SQLiteKeyRepository struct {
	db *sql.DB
}

// NewSQLiteKeyRepository crea un repository de api keys.
//
// Args:
//   - db: conexion SQL compartida.
//
// Returns:
//   - Repository inicializado.
func NewSQLiteKeyRepository(db *sql.DB) *SQLiteKeyRepository {
	return &SQLiteKeyRepository{db: db}
}

// Create inserta una api key.
//
// Args:
//   - ctx: contexto de la operacion.
//   - key: metadata de api key.
//   - hash: hash del secreto.
//
// Returns:
//   - Error si no pudo persistir.
func (r *SQLiteKeyRepository) Create(ctx context.Context, key APIKey, hash string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO api_keys (
			id, owner_id, name, key_hash, key_prefix, scopes, created_at, last_used_at, revoked_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULL, NULL)
	`, key.ID, key.OwnerID, key.Name, hash, key.KeyPrefix, joinScopes(key.Scopes), formatTime(key.CreatedAt))
	if err != nil {
		return fmt.Errorf("crear api key: %w", err)
	}
	return nil
}

// List devuelve api keys activas de un usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//
// Returns:
//   - Lista de api keys sin secretos.
func (r *SQLiteKeyRepository) List(ctx context.Context, ownerID string) ([]APIKey, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, owner_id, name, key_prefix, scopes, created_at, last_used_at, revoked_at
		FROM api_keys
		WHERE owner_id = ? AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, strings.TrimSpace(ownerID))
	if err != nil {
		return nil, fmt.Errorf("listar api keys: %w", err)
	}
	defer rows.Close()
	return scanAPIKeys(rows)
}

// FindActiveByHash busca una api key activa por hash.
//
// Args:
//   - ctx: contexto de la operacion.
//   - hash: hash del secreto recibido.
//
// Returns:
//   - Api key activa o error.
func (r *SQLiteKeyRepository) FindActiveByHash(ctx context.Context, hash string) (APIKey, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, name, key_prefix, scopes, created_at, last_used_at, revoked_at
		FROM api_keys
		WHERE key_hash = ? AND revoked_at IS NULL
	`, strings.TrimSpace(hash))
	return scanAPIKey(row)
}

// TouchLastUsed actualiza la fecha de ultimo uso.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de api key.
//   - usedAt: fecha UTC de uso.
//
// Returns:
//   - Error si no pudo actualizar.
func (r *SQLiteKeyRepository) TouchLastUsed(ctx context.Context, id string, usedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE api_keys
		SET last_used_at = ?
		WHERE id = ? AND revoked_at IS NULL
	`, formatTime(usedAt), strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("actualizar uso de api key: %w", err)
	}
	return nil
}

// Revoke marca una api key como revocada.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//   - id: identificador de api key.
//   - revokedAt: fecha UTC de revocacion.
//
// Returns:
//   - Error si no existe o no pudo revocar.
func (r *SQLiteKeyRepository) Revoke(ctx context.Context, ownerID string, id string, revokedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE api_keys
		SET revoked_at = ?
		WHERE id = ? AND owner_id = ? AND revoked_at IS NULL
	`, formatTime(revokedAt), strings.TrimSpace(id), strings.TrimSpace(ownerID))
	if err != nil {
		return fmt.Errorf("revocar api key: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("confirmar revocacion: %w", err)
	}
	if rows == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}
