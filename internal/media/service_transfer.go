package media

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// CreateTransfer sube archivos, crea envio y genera share.
//
// Args:
//   - ctx: contexto de la operacion.
//   - headers: archivos multipart.
//   - input: datos del envio.
//
// Returns:
//   - Resultado con link corto o error.
func (s *Service) CreateTransfer(ctx context.Context, headers []*multipart.FileHeader, input TransferInput) (TransferResult, error) {
	upload, err := s.UploadFiles(ctx, headers, input.UploadInput)
	if err != nil {
		return TransferResult{}, err
	}
	fileIDs := make([]string, 0, len(upload.Files))
	for _, file := range upload.Files {
		fileIDs = append(fileIDs, file.ID)
	}

	now := time.Now().UTC()
	transfer := Transfer{
		ID:        newTransferID(),
		OwnerID:   input.OwnerID,
		Title:     sanitizeText(input.Title, 160),
		Message:   sanitizeText(input.Message, 600),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if transfer.Title == "" && len(upload.Files) > 0 {
		transfer.Title = upload.Files[0].OriginalName
	}
	if err := s.repo.CreateTransfer(ctx, transfer, fileIDs); err != nil {
		s.rollbackCreated(ctx, upload.Files)
		return TransferResult{}, appError("transfer_create_failed", "No se pudo crear el envio", err)
	}

	share, err := s.CreateShare(ctx, input.OwnerID, false, ShareTargetTransfer, transfer.ID, input.ExpiresDays, input.NeverExpires)
	if err != nil {
		return TransferResult{}, err
	}
	return TransferResult{Transfer: transfer, Share: share, Files: upload.Files, DeletedFiles: upload.DeletedFiles}, nil
}

// ListTransfers devuelve envios visibles.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//   - limit: cantidad maxima.
//
// Returns:
//   - Envios o error.
func (s *Service) ListTransfers(ctx context.Context, ownerID string, includeAll bool, limit int) ([]Transfer, error) {
	transfers, err := s.repo.ListTransfers(ctx, TransferFilter{OwnerID: ownerID, IncludeAll: includeAll, Limit: limit})
	if err != nil {
		return nil, appError("transfer_list_failed", "No se pudieron listar envios", err)
	}
	return transfers, nil
}

// GetTransfer devuelve un envio validando owner.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de envio.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//
// Returns:
//   - Envio o error.
func (s *Service) GetTransfer(ctx context.Context, id string, ownerID string, includeAll bool) (Transfer, error) {
	transfer, err := s.repo.FindTransferByID(ctx, id)
	if err != nil {
		return Transfer{}, appError("transfer_not_found", "Envio no encontrado", ErrNotFound)
	}
	if !includeAll && transfer.OwnerID != ownerID {
		return Transfer{}, appError("transfer_forbidden", "Envio sin permisos", ErrForbidden)
	}
	return transfer, nil
}

func newTransferID() string {
	return uuid.NewString()
}
