package media

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/tomasacebal/go-media-api/internal/storage"
)

// UploadFiles valida cuota y guarda archivos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - headers: archivos multipart recibidos.
//   - input: datos de owner y metadata.
//
// Returns:
//   - Archivos creados, borrados por FIFO o error.
func (s *Service) UploadFiles(ctx context.Context, headers []*multipart.FileHeader, input UploadInput) (UploadResult, error) {
	validated, err := s.validateUpload(headers, input)
	if err != nil {
		return UploadResult{}, err
	}
	deleted, err := s.prepareQuota(ctx, input.OwnerID, validated.totalBytes(), input.ConfirmFIFO)
	if err != nil {
		return UploadResult{}, err
	}

	created := make([]File, 0, len(validated))
	for _, item := range validated {
		file, err := s.saveValidatedFile(ctx, item, input)
		if err != nil {
			s.rollbackCreated(ctx, created)
			return UploadResult{}, err
		}
		created = append(created, file)
	}
	return UploadResult{Files: created, DeletedFiles: deleted}, nil
}

func (s *Service) validateUpload(headers []*multipart.FileHeader, input UploadInput) (validatedFiles, error) {
	if len(headers) == 0 {
		return nil, appError("media_file_required", "El campo file es obligatorio", ErrInvalidUpload)
	}
	if strings.TrimSpace(input.OwnerID) == "" {
		return nil, appError("media_owner_required", "Usuario requerido", ErrInvalidUpload)
	}

	files := make(validatedFiles, 0, len(headers))
	for _, header := range headers {
		if header == nil {
			return nil, appError("media_file_required", "El campo file es obligatorio", ErrInvalidUpload)
		}
		if header.Size <= 0 {
			return nil, appError("media_empty_file", "El archivo esta vacio", ErrFileEmpty)
		}
		if s.maxUploadBytes > 0 && header.Size > s.maxUploadBytes {
			return nil, appError("media_too_large", "El archivo supera el tamaño maximo permitido", ErrFileTooLarge)
		}
		src, err := header.Open()
		if err != nil {
			return nil, appError("media_open_failed", "No se pudo leer el archivo", err)
		}
		mimeType, err := detectMIME(src)
		_ = src.Close()
		if err != nil {
			return nil, err
		}
		files = append(files, validatedFile{
			header:    header,
			mimeType:  mimeType,
			extension: sanitizeExtension(filepath.Ext(header.Filename)),
		})
	}
	return files, nil
}

func (s *Service) saveValidatedFile(ctx context.Context, item validatedFile, input UploadInput) (File, error) {
	src, err := item.header.Open()
	if err != nil {
		return File{}, appError("media_open_failed", "No se pudo leer el archivo", err)
	}
	defer src.Close()

	visibility, err := normalizeVisibility(input.Visibility)
	if err != nil {
		return File{}, appError("media_invalid_visibility", "visibility debe ser public o private", err)
	}
	id := newFileID()
	storedName := id + item.extension
	storagePath := filepath.ToSlash(filepath.Join("users", input.OwnerID, storedName))
	now := time.Now().UTC()
	file := File{
		ID:            id,
		OwnerID:       input.OwnerID,
		OriginalName:  sanitizeFilename(item.header.Filename),
		StoredName:    storedName,
		MIMEType:      item.mimeType,
		Extension:     strings.TrimPrefix(item.extension, "."),
		SizeBytes:     item.header.Size,
		StorageDriver: StorageDriverLocal,
		StoragePath:   storagePath,
		Visibility:    visibility,
		Title:         sanitizeText(input.Title, 160),
		Description:   sanitizeText(input.Description, 600),
		Category:      sanitizeText(input.Category, 120),
		CreatedBy:     sanitizeText(input.CreatedBy, 120),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	saved, err := s.storageProvider.Save(ctx, src, storage.FileMetadata{
		StoredName:  storedName,
		StoragePath: file.StoragePath,
		MIMEType:    file.MIMEType,
		SizeBytes:   file.SizeBytes,
	})
	if err != nil {
		return File{}, appError("media_storage_failed", "No se pudo guardar el archivo", err)
	}
	file.StoragePath = saved.StoragePath
	file.SizeBytes = saved.SizeBytes

	if err := s.repo.CreateFile(ctx, file); err != nil {
		_ = s.storageProvider.Delete(ctx, file.StoragePath)
		return File{}, appError("media_metadata_failed", "No se pudo guardar la metadata", err)
	}
	if err := s.usage.AddStorageUsed(ctx, input.OwnerID, file.SizeBytes); err != nil {
		_ = s.repo.SoftDeleteFile(ctx, file.ID, DeleteReasonUser, time.Now().UTC())
		_ = s.storageProvider.Delete(ctx, file.StoragePath)
		return File{}, appError("quota_update_failed", "No se pudo actualizar la cuota", err)
	}
	return file, nil
}

type validatedFile struct {
	header    *multipart.FileHeader
	mimeType  string
	extension string
}

type validatedFiles []validatedFile

func (files validatedFiles) totalBytes() int64 {
	total := int64(0)
	for _, file := range files {
		total += file.header.Size
	}
	return total
}
