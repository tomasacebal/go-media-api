package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CreateTransfer inserta un envio y sus archivos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - transfer: envio a persistir.
//   - fileIDs: archivos asociados.
//
// Returns:
//   - Error si no pudo persistir.
func (r *SQLiteRepository) CreateTransfer(ctx context.Context, transfer Transfer, fileIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar transfer: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO transfers (id, owner_id, title, message, created_at, updated_at, revoked_at)
		VALUES (?, ?, ?, ?, ?, ?, NULL)
	`, transfer.ID, transfer.OwnerID, transfer.Title, transfer.Message,
		formatMediaTime(transfer.CreatedAt), formatMediaTime(transfer.UpdatedAt)); err != nil {
		return fmt.Errorf("crear transfer: %w", err)
	}
	for _, fileID := range fileIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO transfer_files (transfer_id, file_id, created_at)
			VALUES (?, ?, ?)
		`, transfer.ID, strings.TrimSpace(fileID), formatMediaTime(transfer.CreatedAt)); err != nil {
			return fmt.Errorf("asociar archivo transfer: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar transfer: %w", err)
	}
	return nil
}

// ListTransfers devuelve envios activos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - filter: filtro por owner y limite.
//
// Returns:
//   - Envios o error.
func (r *SQLiteRepository) ListTransfers(ctx context.Context, filter TransferFilter) ([]Transfer, error) {
	query := transferSelectSQL() + " WHERE revoked_at IS NULL"
	args := []interface{}{}
	if !filter.IncludeAll {
		query += " AND owner_id = ?"
		args = append(args, strings.TrimSpace(filter.OwnerID))
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, clampLimit(filter.Limit))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listar transfers: %w", err)
	}
	defer rows.Close()
	return scanTransfers(rows)
}

// FindTransferByID busca un envio activo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de envio.
//
// Returns:
//   - Envio o error.
func (r *SQLiteRepository) FindTransferByID(ctx context.Context, id string) (Transfer, error) {
	row := r.db.QueryRowContext(ctx, transferSelectSQL()+" WHERE id = ? AND revoked_at IS NULL", strings.TrimSpace(id))
	return scanTransfer(row)
}

// ListFilesByTransferID lista archivos activos de un envio.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de envio.
//
// Returns:
//   - Archivos activos o error.
func (r *SQLiteRepository) ListFilesByTransferID(ctx context.Context, id string) ([]File, error) {
	rows, err := r.db.QueryContext(ctx, fileSelectSQL()+`
		INNER JOIN transfer_files ON transfer_files.file_id = media_files.id
		WHERE transfer_files.transfer_id = ? AND media_files.deleted_at IS NULL
		ORDER BY transfer_files.created_at ASC
	`, strings.TrimSpace(id))
	if err != nil {
		return nil, fmt.Errorf("listar archivos transfer: %w", err)
	}
	defer rows.Close()
	return scanFiles(rows)
}

// RemoveFileFromTransfers quita un archivo de todos los envios.
//
// Args:
//   - ctx: contexto de la operacion.
//   - fileID: identificador de archivo.
//
// Returns:
//   - Error si no pudo actualizar.
func (r *SQLiteRepository) RemoveFileFromTransfers(ctx context.Context, fileID string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM transfer_files WHERE file_id = ?", strings.TrimSpace(fileID))
	if err != nil {
		return fmt.Errorf("quitar archivo de transfers: %w", err)
	}
	return nil
}

// RevokeEmptyTransfers revoca envios sin archivos activos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//   - revokedAt: fecha UTC.
//
// Returns:
//   - Error si no pudo revocar.
func (r *SQLiteRepository) RevokeEmptyTransfers(ctx context.Context, ownerID string, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE transfers
		SET revoked_at = ?, updated_at = ?
		WHERE owner_id = ? AND revoked_at IS NULL
			AND NOT EXISTS (
				SELECT 1 FROM transfer_files
				INNER JOIN media_files ON media_files.id = transfer_files.file_id
				WHERE transfer_files.transfer_id = transfers.id AND media_files.deleted_at IS NULL
			)
	`, formatMediaTime(revokedAt), formatMediaTime(revokedAt), strings.TrimSpace(ownerID))
	if err != nil {
		return fmt.Errorf("revocar transfers vacios: %w", err)
	}
	return nil
}

type transferScanner interface {
	Scan(dest ...interface{}) error
}

type transferRows interface {
	transferScanner
	Next() bool
	Err() error
}

func transferSelectSQL() string {
	return `SELECT id, owner_id, title, message, created_at, updated_at, revoked_at FROM transfers`
}

func scanTransfers(rows transferRows) ([]Transfer, error) {
	transfers := make([]Transfer, 0)
	for rows.Next() {
		transfer, err := scanTransfer(rows)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar transfers: %w", err)
	}
	return transfers, nil
}

func scanTransfer(row transferScanner) (Transfer, error) {
	var transfer Transfer
	var createdAt string
	var updatedAt string
	var revokedAt sql.NullString
	err := row.Scan(&transfer.ID, &transfer.OwnerID, &transfer.Title, &transfer.Message, &createdAt, &updatedAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Transfer{}, ErrNotFound
	}
	if err != nil {
		return Transfer{}, fmt.Errorf("leer transfer: %w", err)
	}
	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Transfer{}, fmt.Errorf("parsear created_at transfer: %w", err)
	}
	parsedUpdated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return Transfer{}, fmt.Errorf("parsear updated_at transfer: %w", err)
	}
	transfer.CreatedAt = parsedCreated
	transfer.UpdatedAt = parsedUpdated
	if revokedAt.Valid {
		parsedRevoked, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err != nil {
			return Transfer{}, fmt.Errorf("parsear revoked_at transfer: %w", err)
		}
		transfer.RevokedAt = &parsedRevoked
	}
	return transfer, nil
}
