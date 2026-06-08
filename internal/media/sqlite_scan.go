package media

import (
	"database/sql"
	"fmt"
	"time"
)

type fileRows interface {
	fileScanner
	Next() bool
	Err() error
}

func scanFiles(rows fileRows) ([]File, error) {
	files := make([]File, 0)
	for rows.Next() {
		file, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar metadata: %w", err)
	}
	return files, nil
}

func parseFileTimes(file File, createdAt string, updatedAt string, deletedAt sql.NullString) (File, error) {
	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return File{}, fmt.Errorf("parsear created_at: %w", err)
	}
	parsedUpdated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return File{}, fmt.Errorf("parsear updated_at: %w", err)
	}
	file.CreatedAt = parsedCreated
	file.UpdatedAt = parsedUpdated
	if deletedAt.Valid {
		parsedDeleted, err := time.Parse(time.RFC3339Nano, deletedAt.String)
		if err != nil {
			return File{}, fmt.Errorf("parsear deleted_at: %w", err)
		}
		file.DeletedAt = &parsedDeleted
	}
	return file, nil
}

func formatMediaTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
