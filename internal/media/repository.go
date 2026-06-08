package media

import (
	"context"
	"time"
)

// FileFilter filtra archivos por owner.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Filtro para consultas de archivos.
type FileFilter struct {
	OwnerID    string
	IncludeAll bool
	Limit      int
}

// TransferFilter filtra envios por owner.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Filtro para consultas de envios.
type TransferFilter struct {
	OwnerID    string
	IncludeAll bool
	Limit      int
}

// ShareFilter filtra shares por owner.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Filtro para consultas de shares.
type ShareFilter struct {
	OwnerID    string
	IncludeAll bool
	Limit      int
}

// Repository define persistencia de archivos, envios y shares.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Datos persistidos o errores controlados.
type Repository interface {
	CreateFile(ctx context.Context, file File) error
	ListFiles(ctx context.Context, filter FileFilter) ([]File, error)
	FindFileByID(ctx context.Context, id string) (File, error)
	SoftDeleteFile(ctx context.Context, id string, reason string, deletedAt time.Time) error
	ListFIFOFiles(ctx context.Context, ownerID string) ([]File, error)
	CreateTransfer(ctx context.Context, transfer Transfer, fileIDs []string) error
	ListTransfers(ctx context.Context, filter TransferFilter) ([]Transfer, error)
	FindTransferByID(ctx context.Context, id string) (Transfer, error)
	ListFilesByTransferID(ctx context.Context, id string) ([]File, error)
	RemoveFileFromTransfers(ctx context.Context, fileID string) error
	RevokeEmptyTransfers(ctx context.Context, ownerID string, revokedAt time.Time) error
	CreateShare(ctx context.Context, share Share) error
	ListShares(ctx context.Context, filter ShareFilter) ([]Share, error)
	FindShareByCode(ctx context.Context, code string) (Share, error)
	FindShareByID(ctx context.Context, id string) (Share, error)
	ShareCodeExists(ctx context.Context, code string) (bool, error)
	RevokeShare(ctx context.Context, id string, ownerID string, includeAll bool, revokedAt time.Time) error
	RevokeSharesForFile(ctx context.Context, fileID string, revokedAt time.Time) error
	RevokeSharesForEmptyTransfers(ctx context.Context, ownerID string, revokedAt time.Time) error
}

// UsageRepository define contadores de cuota por usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Cuota, uso y actualizaciones.
type UsageRepository interface {
	GetUsage(ctx context.Context, ownerID string) (int64, int64, error)
	AddStorageUsed(ctx context.Context, ownerID string, delta int64) error
}
