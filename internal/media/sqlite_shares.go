package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CreateShare inserta un link corto.
//
// Args:
//   - ctx: contexto de la operacion.
//   - share: share a persistir.
//
// Returns:
//   - Error si no pudo persistir.
func (r *SQLiteRepository) CreateShare(ctx context.Context, share Share) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO shares (id, owner_id, code, target_type, target_id, expires_at, never_expires, created_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL)
	`, share.ID, share.OwnerID, share.Code, share.TargetType, share.TargetID,
		formatMediaTime(shareStorageExpiry(share)), boolInt(share.NeverExpires), formatMediaTime(share.CreatedAt))
	if err != nil {
		return fmt.Errorf("crear share: %w", err)
	}
	return nil
}

// ListShares devuelve links activos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - filter: filtro por owner y limite.
//
// Returns:
//   - Shares o error.
func (r *SQLiteRepository) ListShares(ctx context.Context, filter ShareFilter) ([]Share, error) {
	query := shareSelectSQL() + " WHERE revoked_at IS NULL"
	args := []interface{}{}
	if !filter.IncludeAll {
		query += " AND owner_id = ?"
		args = append(args, strings.TrimSpace(filter.OwnerID))
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, clampLimit(filter.Limit))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listar shares: %w", err)
	}
	defer rows.Close()
	return scanShares(rows)
}

// FindShareByCode busca un share por codigo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - code: codigo corto.
//
// Returns:
//   - Share o error.
func (r *SQLiteRepository) FindShareByCode(ctx context.Context, code string) (Share, error) {
	row := r.db.QueryRowContext(ctx, shareSelectSQL()+" WHERE code = ? AND revoked_at IS NULL", strings.TrimSpace(code))
	return scanShare(row)
}

// FindShareByID busca un share por id.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador del share.
//
// Returns:
//   - Share o error.
func (r *SQLiteRepository) FindShareByID(ctx context.Context, id string) (Share, error) {
	row := r.db.QueryRowContext(ctx, shareSelectSQL()+" WHERE id = ? AND revoked_at IS NULL", strings.TrimSpace(id))
	return scanShare(row)
}

// ShareCodeExists indica si un codigo ya existe.
//
// Args:
//   - ctx: contexto de la operacion.
//   - code: codigo corto.
//
// Returns:
//   - true si existe o error.
func (r *SQLiteRepository) ShareCodeExists(ctx context.Context, code string) (bool, error) {
	var exists int
	err := r.db.QueryRowContext(ctx, "SELECT 1 FROM shares WHERE code = ? LIMIT 1", strings.TrimSpace(code)).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("buscar codigo share: %w", err)
	}
	return true, nil
}

// RevokeShare revoca un share.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador del share.
//   - ownerID: usuario dueño.
//   - includeAll: permite admin global.
//   - revokedAt: fecha UTC.
//
// Returns:
//   - Error si no existe o no pudo revocar.
func (r *SQLiteRepository) RevokeShare(ctx context.Context, id string, ownerID string, includeAll bool, revokedAt time.Time) error {
	query := "UPDATE shares SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL"
	args := []interface{}{formatMediaTime(revokedAt), strings.TrimSpace(id)}
	if !includeAll {
		query += " AND owner_id = ?"
		args = append(args, strings.TrimSpace(ownerID))
	}
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("revocar share: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("confirmar revocacion share: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeSharesForFile revoca links directos de un archivo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - fileID: archivo borrado.
//   - revokedAt: fecha UTC.
//
// Returns:
//   - Error si no pudo revocar.
func (r *SQLiteRepository) RevokeSharesForFile(ctx context.Context, fileID string, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE shares
		SET revoked_at = ?
		WHERE target_type = ? AND target_id = ? AND revoked_at IS NULL
	`, formatMediaTime(revokedAt), ShareTargetFile, strings.TrimSpace(fileID))
	if err != nil {
		return fmt.Errorf("revocar shares archivo: %w", err)
	}
	return nil
}

// RevokeSharesForEmptyTransfers revoca links de envios vacios.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//   - revokedAt: fecha UTC.
//
// Returns:
//   - Error si no pudo revocar.
func (r *SQLiteRepository) RevokeSharesForEmptyTransfers(ctx context.Context, ownerID string, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE shares
		SET revoked_at = ?
		WHERE owner_id = ? AND target_type = ? AND revoked_at IS NULL
			AND target_id IN (
				SELECT id FROM transfers
				WHERE owner_id = ? AND revoked_at IS NULL
					AND NOT EXISTS (
						SELECT 1 FROM transfer_files
						INNER JOIN media_files ON media_files.id = transfer_files.file_id
						WHERE transfer_files.transfer_id = transfers.id AND media_files.deleted_at IS NULL
					)
			)
	`, formatMediaTime(revokedAt), strings.TrimSpace(ownerID), ShareTargetTransfer, strings.TrimSpace(ownerID))
	if err != nil {
		return fmt.Errorf("revocar shares transfers vacios: %w", err)
	}
	return nil
}

func shareSelectSQL() string {
	return `SELECT id, owner_id, code, target_type, target_id, expires_at, never_expires, created_at, revoked_at FROM shares`
}

func shareStorageExpiry(share Share) time.Time {
	if share.ExpiresAt != nil {
		return *share.ExpiresAt
	}
	return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
