package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// GetUsage devuelve cuota y uso actual.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//
// Returns:
//   - Cuota, uso o error.
func (r *SQLiteUserRepository) GetUsage(ctx context.Context, ownerID string) (int64, int64, error) {
	var quota int64
	var used int64
	err := r.db.QueryRowContext(ctx, `
		SELECT quota_bytes, storage_used_bytes
		FROM users
		WHERE id = ?
	`, strings.TrimSpace(ownerID)).Scan(&quota, &used)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, ErrUserNotFound
	}
	if err != nil {
		return 0, 0, fmt.Errorf("leer uso usuario: %w", err)
	}
	return quota, used, nil
}

// AddStorageUsed suma o resta bytes usados.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//   - delta: bytes a sumar o restar.
//
// Returns:
//   - Error si no pudo actualizar.
func (r *SQLiteUserRepository) AddStorageUsed(ctx context.Context, ownerID string, delta int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET storage_used_bytes = MAX(storage_used_bytes + ?, 0), updated_at = ?
		WHERE id = ?
	`, delta, formatTime(time.Now().UTC()), strings.TrimSpace(ownerID))
	if err != nil {
		return fmt.Errorf("actualizar uso storage: %w", err)
	}
	return nil
}
