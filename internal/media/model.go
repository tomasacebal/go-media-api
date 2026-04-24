package media

import (
	"io"
	"time"
)

const (
	VisibilityPublic   = "public"
	VisibilityPrivate  = "private"
	StorageDriverLocal = "local"
)

// File representa la metadata persistida de un asset.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Metadata serializable para la API.
type File struct {
	ID            string     `json:"id"`
	OriginalName  string     `json:"original_name"`
	StoredName    string     `json:"stored_name"`
	MIMEType      string     `json:"mime_type"`
	Extension     string     `json:"extension"`
	SizeBytes     int64      `json:"size_bytes"`
	StorageDriver string     `json:"storage_driver"`
	StoragePath   string     `json:"storage_path"`
	PublicURL     string     `json:"public_url"`
	Visibility    string     `json:"visibility"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Category      string     `json:"category"`
	CreatedBy     string     `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}

// UploadInput contiene campos opcionales del formulario multipart.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Datos normalizados para crear metadata.
type UploadInput struct {
	Visibility  string
	Title       string
	Description string
	Category    string
	CreatedBy   string
}

// DownloadFile contiene metadata y stream listo para responder.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Archivo descargable validado por el service.
type DownloadFile struct {
	File   File
	Reader io.ReadCloser
}
