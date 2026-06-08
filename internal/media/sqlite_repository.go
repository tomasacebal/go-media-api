package media

import (
	"database/sql"
)

// SQLiteRepository persiste media en SQLite.
//
// Args:
//   - db: conexion SQL compartida.
//
// Returns:
//   - Repository listo para uso.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository crea un repository sobre una conexion compartida.
//
// Args:
//   - db: conexion SQL compartida.
//
// Returns:
//   - Repository SQLite.
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}
