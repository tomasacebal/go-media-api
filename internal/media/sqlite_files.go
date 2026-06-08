package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// CreateFile inserta metadata de archivo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - file: metadata a persistir.
//
// Returns:
//   - Error si no pudo persistir.
func (r *SQLiteRepository) CreateFile(ctx context.Context, file File) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO media_files (
			id, owner_id, original_name, stored_name, mime_type, extension, size_bytes,
			storage_driver, storage_path, public_url, visibility, title, description,
			category, created_by, created_at, updated_at, deleted_at, deleted_reason
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, '')
	`, file.ID, file.OwnerID, file.OriginalName, file.StoredName, file.MIMEType,
		file.Extension, file.SizeBytes, file.StorageDriver, file.StoragePath,
		file.PublicURL, file.Visibility, file.Title, file.Description, file.Category,
		file.CreatedBy, formatMediaTime(file.CreatedAt), formatMediaTime(file.UpdatedAt))
	if err != nil {
		return fmt.Errorf("crear metadata: %w", err)
	}
	return nil
}

// ListFiles devuelve archivos activos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - filter: filtro por owner y limite.
//
// Returns:
//   - Lista de archivos o error.
func (r *SQLiteRepository) ListFiles(ctx context.Context, filter FileFilter) ([]File, error) {
	limit := clampLimit(filter.Limit)
	query := fileSelectSQL() + " WHERE deleted_at IS NULL"
	args := []interface{}{}
	if !filter.IncludeAll {
		query += " AND owner_id = ?"
		args = append(args, strings.TrimSpace(filter.OwnerID))
	}
	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listar metadata: %w", err)
	}
	defer rows.Close()
	return scanFiles(rows)
}

// FindFileByID devuelve metadata activa por id.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador publico.
//
// Returns:
//   - Metadata encontrada o error.
func (r *SQLiteRepository) FindFileByID(ctx context.Context, id string) (File, error) {
	row := r.db.QueryRowContext(ctx, fileSelectSQL()+" WHERE id = ? AND deleted_at IS NULL", strings.TrimSpace(id))
	return scanFile(row)
}

// SoftDeleteFile marca una metadata como borrada.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador publico.
//   - reason: razon de borrado.
//   - deletedAt: fecha UTC.
//
// Returns:
//   - Error si no encontro o no pudo borrar.
func (r *SQLiteRepository) SoftDeleteFile(ctx context.Context, id string, reason string, deletedAt time.Time) error {
	now := formatMediaTime(deletedAt)
	result, err := r.db.ExecContext(ctx, `
		UPDATE media_files
		SET deleted_at = ?, updated_at = ?, deleted_reason = ?
		WHERE id = ? AND deleted_at IS NULL
	`, now, now, strings.TrimSpace(reason), strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("borrar metadata: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("confirmar borrado: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// ListFIFOFiles devuelve archivos activos del mas viejo al mas nuevo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//
// Returns:
//   - Archivos candidatos a limpieza FIFO.
func (r *SQLiteRepository) ListFIFOFiles(ctx context.Context, ownerID string) ([]File, error) {
	rows, err := r.db.QueryContext(ctx, fileSelectSQL()+`
		WHERE owner_id = ? AND deleted_at IS NULL
		ORDER BY created_at ASC
	`, strings.TrimSpace(ownerID))
	if err != nil {
		return nil, fmt.Errorf("listar fifo: %w", err)
	}
	defer rows.Close()
	return scanFiles(rows)
}

type fileScanner interface {
	Scan(dest ...interface{}) error
}

func fileSelectSQL() string {
	return `SELECT media_files.id, media_files.owner_id, media_files.original_name,
		media_files.stored_name, media_files.mime_type, media_files.extension,
		media_files.size_bytes, media_files.storage_driver, media_files.storage_path,
		media_files.public_url, media_files.visibility, media_files.title,
		media_files.description, media_files.category, media_files.created_by,
		media_files.created_at, media_files.updated_at, media_files.deleted_at,
		media_files.deleted_reason FROM media_files`
}

func scanFile(row fileScanner) (File, error) {
	var file File
	var createdAt string
	var updatedAt string
	var deletedAt sql.NullString
	err := row.Scan(&file.ID, &file.OwnerID, &file.OriginalName, &file.StoredName,
		&file.MIMEType, &file.Extension, &file.SizeBytes, &file.StorageDriver,
		&file.StoragePath, &file.PublicURL, &file.Visibility, &file.Title,
		&file.Description, &file.Category, &file.CreatedBy, &createdAt,
		&updatedAt, &deletedAt, &file.DeletedReason)
	if errors.Is(err, sql.ErrNoRows) {
		return File{}, ErrNotFound
	}
	if err != nil {
		return File{}, fmt.Errorf("leer metadata: %w", err)
	}
	return parseFileTimes(file, createdAt, updatedAt, deletedAt)
}
