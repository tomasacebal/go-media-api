package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sort"
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
	List(ctx context.Context) ([]APIKey, error)
	FindActiveByHash(ctx context.Context, hash string) (APIKey, error)
	TouchLastUsed(ctx context.Context, id string, usedAt time.Time) error
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
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
//   - name: nombre visible.
//   - scopes: scopes solicitados.
//
// Returns:
//   - Api key con `secret` visible una sola vez o error.
func (s *APIKeyService) Create(ctx context.Context, name string, scopes []string) (APIKey, error) {
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
		Name:      sanitizeName(name),
		KeyPrefix: secret[:min(len(secret), 16)],
		Scopes:    normalizedScopes,
		CreatedAt: now,
		Secret:    secret,
	}
	if key.Name == "" {
		key.Name = "api key"
	}

	if err := s.repo.Create(ctx, key, hashSecret(secret)); err != nil {
		return APIKey{}, err
	}

	return key, nil
}

// List devuelve api keys activas.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Lista de api keys sin secretos.
func (s *APIKeyService) List(ctx context.Context) ([]APIKey, error) {
	return s.repo.List(ctx)
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

// Revoke revoca una api key activa.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de api key.
//
// Returns:
//   - Error si no existe o no pudo revocar.
func (s *APIKeyService) Revoke(ctx context.Context, id string) error {
	return s.repo.Revoke(ctx, strings.TrimSpace(id), time.Now().UTC())
}

// NormalizeScopes valida y ordena scopes.
//
// Args:
//   - scopes: scopes recibidos.
//
// Returns:
//   - Scopes normalizados o error.
func NormalizeScopes(scopes []string) ([]string, error) {
	allowed := map[string]bool{
		ScopeRead:   true,
		ScopeWrite:  true,
		ScopeDelete: true,
	}
	unique := make(map[string]bool)
	for _, scope := range scopes {
		scope = strings.ToLower(strings.TrimSpace(scope))
		if scope == "" {
			continue
		}
		if !allowed[scope] {
			return nil, ErrInvalidScope
		}
		unique[scope] = true
	}
	if len(unique) == 0 {
		return nil, ErrInvalidScope
	}

	normalized := make([]string, 0, len(unique))
	for scope := range unique {
		normalized = append(normalized, scope)
	}
	sort.Strings(normalized)
	return normalized, nil
}

// HasScope indica si una key contiene un scope.
//
// Args:
//   - scopes: scopes de la key.
//   - required: scope requerido.
//
// Returns:
//   - true si esta permitido.
func HasScope(scopes []string, required string) bool {
	required = strings.ToLower(strings.TrimSpace(required))
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}

func generateSecret() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return keyPrefix + base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func sanitizeName(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 120 {
		return value
	}
	return value[:120]
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
