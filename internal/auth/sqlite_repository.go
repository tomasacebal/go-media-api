package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteKeyRepository persiste api keys en SQLite.
//
// Args:
//   - db: conexion SQL inicializada.
//
// Returns:
//   - Repository listo para uso.
type SQLiteKeyRepository struct {
	db *sql.DB
}

// NewSQLiteKeyRepository abre SQLite y migra api keys.
//
// Args:
//   - dbPath: path del archivo SQLite.
//
// Returns:
//   - Repository SQLite y funcion de cierre.
func NewSQLiteKeyRepository(dbPath string) (*SQLiteKeyRepository, func() error, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, nil, fmt.Errorf("crear carpeta sqlite auth: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("abrir sqlite auth: %w", err)
	}

	repo := &SQLiteKeyRepository{db: db}
	if err := repo.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, nil, err
	}

	return repo, db.Close, nil
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
			id, name, key_hash, key_prefix, scopes, created_at, last_used_at, revoked_at
		) VALUES (?, ?, ?, ?, ?, ?, NULL, NULL)
	`,
		key.ID,
		key.Name,
		hash,
		key.KeyPrefix,
		joinScopes(key.Scopes),
		key.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("crear api key: %w", err)
	}

	return nil
}

// List devuelve api keys activas.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Lista de api keys sin secretos.
func (r *SQLiteKeyRepository) List(ctx context.Context) ([]APIKey, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, key_prefix, scopes, created_at, last_used_at, revoked_at
		FROM api_keys
		WHERE revoked_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("listar api keys: %w", err)
	}
	defer rows.Close()

	keys := make([]APIKey, 0)
	for rows.Next() {
		key, err := scanAPIKey(rows)
		if err != nil {
			return nil, fmt.Errorf("leer api key: %w", err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar api keys: %w", err)
	}

	return keys, nil
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
		SELECT id, name, key_prefix, scopes, created_at, last_used_at, revoked_at
		FROM api_keys
		WHERE key_hash = ? AND revoked_at IS NULL
	`, hash)

	key, err := scanAPIKey(row)
	if errors.Is(err, sql.ErrNoRows) {
		return APIKey{}, ErrAPIKeyNotFound
	}
	if err != nil {
		return APIKey{}, fmt.Errorf("leer api key: %w", err)
	}

	return key, nil
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
	`, usedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("actualizar uso de api key: %w", err)
	}
	return nil
}

// Revoke marca una api key como revocada.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de api key.
//   - revokedAt: fecha UTC de revocacion.
//
// Returns:
//   - Error si no existe o no pudo revocar.
func (r *SQLiteKeyRepository) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE api_keys
		SET revoked_at = ?
		WHERE id = ? AND revoked_at IS NULL
	`, revokedAt.UTC().Format(time.RFC3339Nano), id)
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

func (r *SQLiteKeyRepository) migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key_hash TEXT NOT NULL UNIQUE,
			key_prefix TEXT NOT NULL,
			scopes TEXT NOT NULL,
			created_at TEXT NOT NULL,
			last_used_at TEXT,
			revoked_at TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);
		CREATE INDEX IF NOT EXISTS idx_api_keys_revoked_at ON api_keys(revoked_at);
	`)
	if err != nil {
		return fmt.Errorf("migrar sqlite auth: %w", err)
	}

	return nil
}

type apiKeyScanner interface {
	Scan(dest ...any) error
}

func scanAPIKey(row apiKeyScanner) (APIKey, error) {
	var key APIKey
	var scopes string
	var createdAt string
	var lastUsedAt sql.NullString
	var revokedAt sql.NullString

	err := row.Scan(
		&key.ID,
		&key.Name,
		&key.KeyPrefix,
		&scopes,
		&createdAt,
		&lastUsedAt,
		&revokedAt,
	)
	if err != nil {
		return APIKey{}, err
	}

	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return APIKey{}, fmt.Errorf("parsear created_at: %w", err)
	}
	key.CreatedAt = parsedCreatedAt
	key.Scopes = splitScopes(scopes)

	if lastUsedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, lastUsedAt.String)
		if err != nil {
			return APIKey{}, fmt.Errorf("parsear last_used_at: %w", err)
		}
		key.LastUsedAt = &parsed
	}
	if revokedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err != nil {
			return APIKey{}, fmt.Errorf("parsear revoked_at: %w", err)
		}
		key.RevokedAt = &parsed
	}

	return key, nil
}

func joinScopes(scopes []string) string {
	return strings.Join(scopes, ",")
}

func splitScopes(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	scopes := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			scopes = append(scopes, part)
		}
	}
	return scopes
}
