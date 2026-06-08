package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type apiKeyScanner interface {
	Scan(dest ...interface{}) error
}

type apiKeyRows interface {
	apiKeyScanner
	Next() bool
	Err() error
}

func scanAPIKeys(rows apiKeyRows) ([]APIKey, error) {
	keys := make([]APIKey, 0)
	for rows.Next() {
		key, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar api keys: %w", err)
	}
	return keys, nil
}

func scanAPIKey(row apiKeyScanner) (APIKey, error) {
	var key APIKey
	var scopes string
	var createdAt string
	var lastUsedAt sql.NullString
	var revokedAt sql.NullString

	err := row.Scan(&key.ID, &key.OwnerID, &key.Name, &key.KeyPrefix, &scopes, &createdAt, &lastUsedAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return APIKey{}, ErrAPIKeyNotFound
	}
	if err != nil {
		return APIKey{}, fmt.Errorf("leer api key: %w", err)
	}

	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return APIKey{}, fmt.Errorf("parsear created_at api key: %w", err)
	}
	key.CreatedAt = parsedCreatedAt
	key.Scopes = splitScopes(scopes)
	if lastUsedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, lastUsedAt.String)
		if err != nil {
			return APIKey{}, fmt.Errorf("parsear last_used_at api key: %w", err)
		}
		key.LastUsedAt = &parsed
	}
	if revokedAt.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err != nil {
			return APIKey{}, fmt.Errorf("parsear revoked_at api key: %w", err)
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
