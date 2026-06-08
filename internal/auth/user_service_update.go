package auth

import (
	"context"
	"strings"
	"time"
)

// Update aplica cambios administrativos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de usuario.
//   - patch: cambios solicitados.
//
// Returns:
//   - Usuario actualizado o error.
func (s *UserService) Update(ctx context.Context, id string, patch UserPatch) (User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	if patch.Name != nil {
		user.Name = sanitizeUserText(*patch.Name, 120)
	}
	if patch.Role != nil {
		user.Role = normalizeRole(*patch.Role)
	}
	if patch.QuotaBytes != nil && *patch.QuotaBytes > 0 {
		user.QuotaBytes = *patch.QuotaBytes
	}
	if patch.ShareTTLDays != nil && *patch.ShareTTLDays >= 0 {
		user.ShareTTLDays = *patch.ShareTTLDays
	}
	if patch.Disabled != nil {
		now := time.Now().UTC()
		if *patch.Disabled {
			user.DisabledAt = &now
		} else {
			user.DisabledAt = nil
		}
	}
	user.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, user); err != nil {
		return User{}, err
	}
	return user, nil
}

// ChangePassword actualiza el password de una cuenta.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: usuario autenticado.
//   - currentPassword: password actual.
//   - nextPassword: password nuevo.
//
// Returns:
//   - Error si no pudo cambiarlo.
func (s *UserService) ChangePassword(ctx context.Context, id string, currentPassword string, nextPassword string) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if !VerifyPassword(currentPassword, user.PasswordHash) || len(nextPassword) < 8 {
		return ErrInvalidUserInput
	}
	hash, err := HashPassword(nextPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, id, hash)
}

// GetUsage devuelve cuota y uso del usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//
// Returns:
//   - Cuota, uso o error.
func (s *UserService) GetUsage(ctx context.Context, ownerID string) (int64, int64, error) {
	return s.repo.GetUsage(ctx, ownerID)
}

// AddStorageUsed actualiza el contador de uso.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//   - delta: bytes a sumar o restar.
//
// Returns:
//   - Error si no pudo actualizar.
func (s *UserService) AddStorageUsed(ctx context.Context, ownerID string, delta int64) error {
	return s.repo.AddStorageUsed(ctx, ownerID, delta)
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeRole(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == RoleAdmin {
		return RoleAdmin
	}
	return RoleUser
}

func sanitizeUserText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) > limit {
		return value[:limit]
	}
	return value
}
