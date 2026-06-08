package media

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type shareScanner interface {
	Scan(dest ...interface{}) error
}

type shareRows interface {
	shareScanner
	Next() bool
	Err() error
}

func scanShares(rows shareRows) ([]Share, error) {
	shares := make([]Share, 0)
	for rows.Next() {
		share, err := scanShare(rows)
		if err != nil {
			return nil, err
		}
		shares = append(shares, share)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar shares: %w", err)
	}
	return shares, nil
}

func scanShare(row shareScanner) (Share, error) {
	var share Share
	var expiresAt string
	var neverExpires int
	var createdAt string
	var revokedAt sql.NullString
	err := row.Scan(&share.ID, &share.OwnerID, &share.Code, &share.TargetType, &share.TargetID, &expiresAt, &neverExpires, &createdAt, &revokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Share{}, ErrNotFound
	}
	if err != nil {
		return Share{}, fmt.Errorf("leer share: %w", err)
	}
	parsedExpires, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return Share{}, fmt.Errorf("parsear expires_at share: %w", err)
	}
	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Share{}, fmt.Errorf("parsear created_at share: %w", err)
	}
	share.NeverExpires = neverExpires == 1
	if !share.NeverExpires {
		share.ExpiresAt = &parsedExpires
	}
	share.CreatedAt = parsedCreated
	if revokedAt.Valid {
		parsedRevoked, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err != nil {
			return Share{}, fmt.Errorf("parsear revoked_at share: %w", err)
		}
		share.RevokedAt = &parsedRevoked
	}
	return share, nil
}
