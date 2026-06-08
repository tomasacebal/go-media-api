package media

import (
	"io"
	"time"
)

const (
	VisibilityPublic       = "public"
	VisibilityPrivate      = "private"
	StorageDriverLocal     = "local"
	ShareTargetFile        = "file"
	ShareTargetTransfer    = "transfer"
	DeleteReasonUser       = "user"
	DeleteReasonQuotaFIFO  = "quota_fifo"
	defaultShareCodeLength = 8
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
	OwnerID       string     `json:"owner_id"`
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
	DeletedReason string     `json:"deleted_reason,omitempty"`
}

// Transfer representa un envio multiarchivo.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Envio serializable.
type Transfer struct {
	ID        string     `json:"id"`
	OwnerID   string     `json:"owner_id"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// Share representa un link corto.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Link compartible serializable.
type Share struct {
	ID           string     `json:"id"`
	OwnerID      string     `json:"owner_id"`
	Code         string     `json:"code"`
	URL          string     `json:"url"`
	TargetType   string     `json:"target_type"`
	TargetID     string     `json:"target_id"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	NeverExpires bool       `json:"never_expires"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	SharedFiles  []string   `json:"shared_files,omitempty"`
}

// UploadInput contiene campos opcionales del upload.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Datos normalizados para crear metadata.
type UploadInput struct {
	OwnerID     string
	Visibility  string
	Title       string
	Description string
	Category    string
	CreatedBy   string
	ConfirmFIFO bool
}

// TransferInput contiene datos para crear un envio.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Datos de envio y share.
type TransferInput struct {
	UploadInput
	Message      string
	ExpiresDays  int
	NeverExpires bool
}

// UploadResult describe el resultado de guardar archivos.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Archivos creados y borrados por cuota.
type UploadResult struct {
	Files        []File `json:"files"`
	DeletedFiles []File `json:"deleted_files"`
}

// TransferResult describe un envio creado.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Transferencia, share y archivos asociados.
type TransferResult struct {
	Transfer     Transfer `json:"transfer"`
	Share        Share    `json:"share"`
	Files        []File   `json:"files"`
	DeletedFiles []File   `json:"deleted_files"`
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

// ShareDetail describe una pagina publica de share.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Share y archivos visibles publicamente.
type ShareDetail struct {
	Share    Share     `json:"share"`
	Transfer *Transfer `json:"transfer,omitempty"`
	Files    []File    `json:"files"`
}

// QuotaPreview informa limpieza FIFO requerida.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Datos para confirmar limpieza antes de subir.
type QuotaPreview struct {
	QuotaBytes     int64  `json:"quota_bytes"`
	UsedBytes      int64  `json:"used_bytes"`
	UploadBytes    int64  `json:"upload_bytes"`
	BytesToFree    int64  `json:"bytes_to_free"`
	ReclaimedBytes int64  `json:"reclaimed_bytes"`
	Files          []File `json:"files"`
}
