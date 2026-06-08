package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"sort"
	"strings"
)

// NormalizeScopes valida y ordena scopes.
//
// Args:
//   - scopes: scopes recibidos.
//
// Returns:
//   - Scopes normalizados o error.
func NormalizeScopes(scopes []string) ([]string, error) {
	allowed := map[string]bool{ScopeRead: true, ScopeWrite: true, ScopeDelete: true}
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
