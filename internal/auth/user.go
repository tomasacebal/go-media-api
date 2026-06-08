package auth

import "time"

const (
	RoleAdmin        = "admin"
	RoleUser         = "user"
	LocalSessionUser = "session_user"
)

// User representa una cuenta del producto.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Usuario persistido con datos de cuota y rol.
type User struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	Name             string     `json:"name"`
	Role             string     `json:"role"`
	PasswordHash     string     `json:"-"`
	QuotaBytes       int64      `json:"quota_bytes"`
	ShareTTLDays     int        `json:"share_ttl_days"`
	StorageUsedBytes int64      `json:"storage_used_bytes"`
	DisabledAt       *time.Time `json:"disabled_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// SessionUser contiene datos seguros para request locals.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Identidad autenticada sin secretos.
type SessionUser struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Role             string `json:"role"`
	QuotaBytes       int64  `json:"quota_bytes"`
	ShareTTLDays     int    `json:"share_ttl_days"`
	StorageUsedBytes int64  `json:"storage_used_bytes"`
}

// CreateUserInput describe el alta de usuario.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Datos normalizados para crear una cuenta.
type CreateUserInput struct {
	Email              string
	Name               string
	Password           string
	Role               string
	QuotaBytes         int64
	ShareTTLDays       int
	UseDefaultShareTTL bool
}

// UserPatch describe cambios administrativos de usuario.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Campos opcionales para actualizar una cuenta.
type UserPatch struct {
	Name         *string
	Role         *string
	QuotaBytes   *int64
	ShareTTLDays *int
	Disabled     *bool
}

// Session devuelve la version segura del usuario.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Usuario seguro para sesiones y API.
func (u User) Session() SessionUser {
	return SessionUser{
		ID:               u.ID,
		Email:            u.Email,
		Name:             u.Name,
		Role:             u.Role,
		QuotaBytes:       u.QuotaBytes,
		ShareTTLDays:     u.ShareTTLDays,
		StorageUsedBytes: u.StorageUsedBytes,
	}
}

// IsAdmin indica si el usuario tiene rol admin.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - true si el rol es admin.
func (u SessionUser) IsAdmin() bool {
	return u.Role == RoleAdmin
}
