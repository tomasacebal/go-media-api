package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ScopeRead   = "read"
	ScopeWrite  = "write"
	ScopeDelete = "delete"
	keyPrefix   = "gma_"
)

var (
	ErrAPIKeyNotFound  = errors.New("api key no encontrada")
	ErrAPIKeyForbidden = errors.New("api key sin permisos")
	ErrInvalidScope    = errors.New("scope invalido")
)

// APIKey representa una credencial para clientes externos.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Metadata segura para serializar.
type APIKey struct {
	ID         string     `json:"id"`
	OwnerID    string     `json:"owner_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	Secret     string     `json:"secret,omitempty"`
}

// APIKeyRepository define persistencia para api keys.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Api keys activas o errores controlados.
type APIKeyRepository interface {
	Create(ctx context.Context, key APIKey, hash string) error
	List(ctx context.Context, ownerID string) ([]APIKey, error)
	FindActiveByHash(ctx context.Context, hash string) (APIKey, error)
	TouchLastUsed(ctx context.Context, id string, usedAt time.Time) error
	Revoke(ctx context.Context, ownerID string, id string, revokedAt time.Time) error
}

// APIKeyService crea y valida api keys.
//
// Args:
//   - repo: repository de api keys.
//
// Returns:
//   - Servicio listo para handlers y middleware.
type APIKeyService struct {
	repo APIKeyRepository
}

// NewAPIKeyService crea un servicio de api keys.
//
// Args:
//   - repo: repository de api keys.
//
// Returns:
//   - APIKeyService inicializado.
func NewAPIKeyService(repo APIKeyRepository) *APIKeyService {
	return &APIKeyService{repo: repo}
}

// Create genera una api key y persiste solo su hash.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño de la key.
//   - name: nombre visible.
//   - scopes: scopes solicitados.
//
// Returns:
//   - Api key con `secret` visible una sola vez o error.
func (s *APIKeyService) Create(ctx context.Context, ownerID string, name string, scopes []string) (APIKey, error) {
	normalizedScopes, err := NormalizeScopes(scopes)
	if err != nil {
		return APIKey{}, err
	}
	secret, err := generateSecret()
	if err != nil {
		return APIKey{}, err
	}

	now := time.Now().UTC()
	key := APIKey{
		ID:        uuid.NewString(),
		OwnerID:   strings.TrimSpace(ownerID),
		Name:      sanitizeName(name),
		KeyPrefix: secret[:min(len(secret), 16)],
		Scopes:    normalizedScopes,
		CreatedAt: now,
		Secret:    secret,
	}
	if key.Name == "" {
		key.Name = "api key"
	}
	if key.OwnerID == "" {
		return APIKey{}, ErrAPIKeyNotFound
	}
	if err := s.repo.Create(ctx, key, hashSecret(secret)); err != nil {
		return APIKey{}, err
	}
	return key, nil
}

// List devuelve api keys activas de un usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//
// Returns:
//   - Lista de api keys sin secretos.
func (s *APIKeyService) List(ctx context.Context, ownerID string) ([]APIKey, error) {
	return s.repo.List(ctx, strings.TrimSpace(ownerID))
}

// Authenticate valida una api key y su scope.
//
// Args:
//   - ctx: contexto de la operacion.
//   - secret: api key recibida.
//   - scope: scope requerido.
//
// Returns:
//   - Api key validada o error.
func (s *APIKeyService) Authenticate(ctx context.Context, secret string, scope string) (APIKey, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return APIKey{}, ErrAPIKeyNotFound
	}
	key, err := s.repo.FindActiveByHash(ctx, hashSecret(secret))
	if err != nil {
		return APIKey{}, err
	}
	if !HasScope(key.Scopes, scope) {
		return APIKey{}, ErrAPIKeyForbidden
	}

	now := time.Now().UTC()
	_ = s.repo.TouchLastUsed(ctx, key.ID, now)
	key.LastUsedAt = &now
	return key, nil
}

// Revoke revoca una api key activa del usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueño.
//   - id: identificador de api key.
//
// Returns:
//   - Error si no existe o no pudo revocar.
func (s *APIKeyService) Revoke(ctx context.Context, ownerID string, id string) error {
	return s.repo.Revoke(ctx, strings.TrimSpace(ownerID), strings.TrimSpace(id), time.Now().UTC())
}
