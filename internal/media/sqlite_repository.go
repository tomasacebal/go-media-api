package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteRepository persiste metadata en SQLite.
//
// Args:
//   - db: conexion SQL inicializada.
//
// Returns:
//   - Repository listo para uso.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository abre SQLite y ejecuta la migracion minima.
//
// Args:
//   - dbPath: path del archivo SQLite.
//
// Returns:
//   - Repository SQLite y funcion de cierre.
func NewSQLiteRepository(dbPath string) (*SQLiteRepository, func() error, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, nil, fmt.Errorf("crear carpeta sqlite: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("abrir sqlite: %w", err)
	}

	repo := &SQLiteRepository{db: db}
	if err := repo.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	return repo, db.Close, nil
}

// Create inserta metadata de un archivo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - file: metadata a persistir.
//
// Returns:
//   - Error si no pudo persistir.
func (r *SQLiteRepository) Create(ctx context.Context, file File) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO media_files (
			id, original_name, stored_name, mime_type, extension, size_bytes,
			storage_driver, storage_path, public_url, visibility, title,
			description, category, created_by, created_at, updated_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)
	`,
		file.ID,
		file.OriginalName,
		file.StoredName,
		file.MIMEType,
		file.Extension,
		file.SizeBytes,
		file.StorageDriver,
		file.StoragePath,
		file.PublicURL,
		file.Visibility,
		file.Title,
		file.Description,
		file.Category,
		file.CreatedBy,
		file.CreatedAt.UTC().Format(time.RFC3339Nano),
		file.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("crear metadata: %w", err)
	}

	return nil
}

// List devuelve metadata activa ordenada por fecha de creacion.
//
// Args:
//   - ctx: contexto de la operacion.
//   - limit: cantidad maxima de resultados.
//
// Returns:
//   - Lista de metadata activa o error.
func (r *SQLiteRepository) List(ctx context.Context, limit int) ([]File, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, original_name, stored_name, mime_type, extension, size_bytes,
			storage_driver, storage_path, public_url, visibility, title,
			description, category, created_by, created_at, updated_at, deleted_at
		FROM media_files
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("listar metadata: %w", err)
	}
	defer rows.Close()

	files := make([]File, 0)
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, fmt.Errorf("leer fila de metadata: %w", err)
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar metadata: %w", err)
	}

	return files, nil
}

// FindByID devuelve metadata activa por id.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador publico.
//
// Returns:
//   - Metadata encontrada o error.
func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (File, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, original_name, stored_name, mime_type, extension, size_bytes,
			storage_driver, storage_path, public_url, visibility, title,
			description, category, created_by, created_at, updated_at, deleted_at
		FROM media_files
		WHERE id = ? AND deleted_at IS NULL
	`, id)

	file, err := scanFile(row)
	if errors.Is(err, sql.ErrNoRows) {
		return File{}, ErrNotFound
	}
	if err != nil {
		return File{}, fmt.Errorf("leer metadata: %w", err)
	}

	return file, nil
}

// SoftDelete marca una metadata como borrada.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador publico.
//
// Returns:
//   - Error si no encontro o no pudo borrar.
func (r *SQLiteRepository) SoftDelete(ctx context.Context, id string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := r.db.ExecContext(ctx, `
		UPDATE media_files
		SET deleted_at = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`, now, now, id)
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

func (r *SQLiteRepository) migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS media_files (
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
		CREATE INDEX IF NOT EXISTS idx_media_files_visibility ON media_files(visibility);
		CREATE INDEX IF NOT EXISTS idx_media_files_deleted_at ON media_files(deleted_at);
	`)
	if err != nil {
		return fmt.Errorf("migrar sqlite media: %w", err)
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFile(row scanner) (File, error) {
	var file File
	var createdAt string
	var updatedAt string
	var deletedAt sql.NullString

	err := row.Scan(
		&file.ID,
		&file.OriginalName,
		&file.StoredName,
		&file.MIMEType,
		&file.Extension,
		&file.SizeBytes,
		&file.StorageDriver,
		&file.StoragePath,
		&file.PublicURL,
		&file.Visibility,
		&file.Title,
		&file.Description,
		&file.Category,
		&file.CreatedBy,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return File{}, err
	}

	file.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return File{}, fmt.Errorf("parsear created_at: %w", err)
	}
	file.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return File{}, fmt.Errorf("parsear updated_at: %w", err)
	}
	if deletedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, deletedAt.String)
		if err != nil {
			return File{}, fmt.Errorf("parsear deleted_at: %w", err)
		}
		file.DeletedAt = &parsed
	}

	return file, nil
}
