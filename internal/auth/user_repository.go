package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SQLiteUserRepository persiste usuarios en SQLite.
//
// Args:
//   - db: conexion SQL compartida.
//
// Returns:
//   - Repository listo para usuarios.
type SQLiteUserRepository struct {
	db *sql.DB
}

// NewSQLiteUserRepository crea un repository de usuarios.
//
// Args:
//   - db: conexion SQL compartida.
//
// Returns:
//   - Repository inicializado.
func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

// Count devuelve la cantidad de usuarios.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Cantidad de usuarios o error.
func (r *SQLiteUserRepository) Count(ctx context.Context) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return 0, fmt.Errorf("contar usuarios: %w", err)
	}
	return count, nil
}

// Create inserta un usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//   - user: usuario a persistir.
//
// Returns:
//   - Error si no pudo crear.
func (r *SQLiteUserRepository) Create(ctx context.Context, user User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (
			id, email, name, role, password_hash, quota_bytes, share_ttl_days, storage_used_bytes,
			disabled_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)
	`, user.ID, user.Email, user.Name, user.Role, user.PasswordHash,
		user.QuotaBytes, user.ShareTTLDays, user.StorageUsedBytes, formatTime(user.CreatedAt), formatTime(user.UpdatedAt))
	if err != nil {
		return fmt.Errorf("crear usuario: %w", err)
	}
	return nil
}

// List devuelve usuarios ordenados por fecha.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Usuarios persistidos o error.
func (r *SQLiteUserRepository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, email, name, role, password_hash, quota_bytes, share_ttl_days, storage_used_bytes,
			disabled_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("listar usuarios: %w", err)
	}
	defer rows.Close()
	return scanUsers(rows)
}

// FindByID busca un usuario por id.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de usuario.
//
// Returns:
//   - Usuario encontrado o error.
func (r *SQLiteUserRepository) FindByID(ctx context.Context, id string) (User, error) {
	row := r.db.QueryRowContext(ctx, userSelectSQL()+" WHERE id = ?", strings.TrimSpace(id))
	return scanUser(row)
}

// FindByEmail busca un usuario por email.
//
// Args:
//   - ctx: contexto de la operacion.
//   - email: email normalizado.
//
// Returns:
//   - Usuario encontrado o error.
func (r *SQLiteUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	row := r.db.QueryRowContext(ctx, userSelectSQL()+" WHERE lower(email) = lower(?)", strings.TrimSpace(email))
	return scanUser(row)
}

// FirstAdmin devuelve el primer admin creado.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Usuario admin o error.
func (r *SQLiteUserRepository) FirstAdmin(ctx context.Context) (User, error) {
	row := r.db.QueryRowContext(ctx, userSelectSQL()+" WHERE role = ? ORDER BY created_at ASC LIMIT 1", RoleAdmin)
	return scanUser(row)
}

// Update persiste cambios administrativos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - user: usuario con campos actualizados.
//
// Returns:
//   - Error si no pudo actualizar.
func (r *SQLiteUserRepository) Update(ctx context.Context, user User) error {
	var disabledAt *string
	if user.DisabledAt != nil {
		value := formatTime(*user.DisabledAt)
		disabledAt = &value
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET email = ?, name = ?, role = ?, quota_bytes = ?, share_ttl_days = ?, disabled_at = ?, updated_at = ?
		WHERE id = ?
	`, user.Email, user.Name, user.Role, user.QuotaBytes, user.ShareTTLDays, disabledAt, formatTime(user.UpdatedAt), user.ID)
	if err != nil {
		return fmt.Errorf("actualizar usuario: %w", err)
	}
	return nil
}

// UpdatePassword guarda un nuevo hash de password.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de usuario.
//   - passwordHash: hash nuevo.
//
// Returns:
//   - Error si no pudo actualizar.
func (r *SQLiteUserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = ?, updated_at = ?
		WHERE id = ?
	`, passwordHash, formatTime(time.Now().UTC()), strings.TrimSpace(id))
	if err != nil {
		return fmt.Errorf("actualizar password: %w", err)
	}
	return nil
}
