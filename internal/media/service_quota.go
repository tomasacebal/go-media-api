package media

import (
	"context"
	"time"
)

func (s *Service) prepareQuota(ctx context.Context, ownerID string, uploadBytes int64, confirm bool) ([]File, error) {
	quota, used, err := s.usage.GetUsage(ctx, ownerID)
	if err != nil {
		return nil, appError("quota_lookup_failed", "No se pudo leer la cuota", err)
	}
	if quota <= 0 {
		return nil, appError("quota_invalid", "Cuota invalida", ErrQuotaExceeded)
	}
	if uploadBytes > quota {
		return nil, appError("file_exceeds_quota", "El archivo supera la cuota disponible", ErrFileExceedsQuota)
	}
	if used+uploadBytes <= quota {
		return nil, nil
	}

	preview, err := s.buildQuotaPreview(ctx, ownerID, quota, used, uploadBytes)
	if err != nil {
		return nil, err
	}
	if preview.ReclaimedBytes < preview.BytesToFree {
		return nil, appError("quota_exceeded", "No hay espacio suficiente para liberar", ErrQuotaExceeded)
	}
	if !confirm {
		return nil, &QuotaCleanupRequiredError{Preview: preview}
	}
	return s.deleteQuotaFiles(ctx, preview.Files)
}

func (s *Service) buildQuotaPreview(ctx context.Context, ownerID string, quota int64, used int64, uploadBytes int64) (QuotaPreview, error) {
	bytesToFree := used + uploadBytes - quota
	candidates, err := s.repo.ListFIFOFiles(ctx, ownerID)
	if err != nil {
		return QuotaPreview{}, appError("quota_fifo_failed", "No se pudo calcular limpieza", err)
	}
	files := make([]File, 0)
	reclaimed := int64(0)
	for _, file := range candidates {
		files = append(files, file)
		reclaimed += file.SizeBytes
		if reclaimed >= bytesToFree {
			break
		}
	}
	return QuotaPreview{QuotaBytes: quota, UsedBytes: used, UploadBytes: uploadBytes, BytesToFree: bytesToFree, ReclaimedBytes: reclaimed, Files: files}, nil
}

func (s *Service) deleteQuotaFiles(ctx context.Context, files []File) ([]File, error) {
	deleted := make([]File, 0, len(files))
	for _, file := range files {
		if err := s.deleteFile(ctx, file, DeleteReasonQuotaFIFO); err != nil {
			return deleted, err
		}
		deleted = append(deleted, file)
	}
	return deleted, nil
}

func (s *Service) deleteFile(ctx context.Context, file File, reason string) error {
	now := time.Now().UTC()
	if err := s.repo.RevokeSharesForFile(ctx, file.ID, now); err != nil {
		return appError("share_revoke_failed", "No se pudieron revocar links", err)
	}
	if err := s.repo.SoftDeleteFile(ctx, file.ID, reason, now); err != nil {
		return appError("media_delete_failed", "No se pudo borrar la metadata", err)
	}
	if err := s.storageProvider.Delete(ctx, file.StoragePath); err != nil {
		return appError("media_storage_delete_failed", "No se pudo borrar el archivo", err)
	}
	if err := s.usage.AddStorageUsed(ctx, file.OwnerID, -file.SizeBytes); err != nil {
		return appError("quota_update_failed", "No se pudo actualizar la cuota", err)
	}
	if err := s.repo.RemoveFileFromTransfers(ctx, file.ID); err != nil {
		return appError("transfer_update_failed", "No se pudo actualizar el envio", err)
	}
	if err := s.repo.RevokeSharesForEmptyTransfers(ctx, file.OwnerID, now); err != nil {
		return appError("share_revoke_failed", "No se pudieron revocar links vacios", err)
	}
	if err := s.repo.RevokeEmptyTransfers(ctx, file.OwnerID, now); err != nil {
		return appError("transfer_revoke_failed", "No se pudieron revocar envios vacios", err)
	}
	return nil
}

func (s *Service) rollbackCreated(ctx context.Context, files []File) {
	for _, file := range files {
		_ = s.deleteFile(ctx, file, DeleteReasonUser)
	}
}
