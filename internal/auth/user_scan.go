package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrUserNotFound = errors.New("usuario no encontrado")

type userScanner interface {
	Scan(dest ...interface{}) error
}

type userRows interface {
	userScanner
	Next() bool
	Err() error
}

func userSelectSQL() string {
	return `SELECT id, email, name, role, password_hash, quota_bytes, share_ttl_days, storage_used_bytes,
		disabled_at, created_at, updated_at FROM users`
}

func scanUsers(rows userRows) ([]User, error) {
	users := make([]User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar usuarios: %w", err)
	}
	return users, nil
}

func scanUser(row userScanner) (User, error) {
	var user User
	var disabledAt sql.NullString
	var createdAt string
	var updatedAt string

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Role,
		&user.PasswordHash,
		&user.QuotaBytes,
		&user.ShareTTLDays,
		&user.StorageUsedBytes,
		&disabledAt,
		&createdAt,
		&updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("leer usuario: %w", err)
	}

	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("parsear created_at usuario: %w", err)
	}
	parsedUpdated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return User{}, fmt.Errorf("parsear updated_at usuario: %w", err)
	}
	user.CreatedAt = parsedCreated
	user.UpdatedAt = parsedUpdated
	if disabledAt.Valid {
		parsedDisabled, err := time.Parse(time.RFC3339Nano, disabledAt.String)
		if err != nil {
			return User{}, fmt.Errorf("parsear disabled_at usuario: %w", err)
		}
		user.DisabledAt = &parsedDisabled
	}
	return user, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
