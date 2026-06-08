package media

import (
	"context"
	"errors"
	"strings"

	"github.com/tomasacebal/go-media-api/internal/storage"
)

const sniffSize = 512

// Service coordina validacion, storage y metadata de media.
//
// Args:
//   - repo: repository de metadata.
//   - storageProvider: provider de archivos.
//   - usage: repositorio de cuotas.
//   - maxUploadBytes: limite maximo permitido por archivo.
//   - publicBaseURL: URL publica base.
//   - shareTTLDays: dias default de links.
//
// Returns:
//   - Servicio listo para handlers.
type Service struct {
	repo            Repository
	storageProvider storage.Provider
	usage           UsageRepository
	maxUploadBytes  int64
	publicBaseURL   string
	shareTTLDays    int
	storagePath     string
	sessionStore    *ChunkedSessionStore
}

// NewService crea el servicio de media.
//
// Args:
//   - repo: repository de metadata.
//   - storageProvider: provider de archivos.
//   - usage: repositorio de cuotas.
//   - maxUploadBytes: limite maximo permitido.
//   - publicBaseURL: URL publica base.
//   - shareTTLDays: dias default de links.
//   - storagePath: ruta base del almacenamiento local.
//
// Returns:
//   - Servicio inicializado.
func NewService(repo Repository, storageProvider storage.Provider, usage UsageRepository, maxUploadBytes int64, publicBaseURL string, shareTTLDays int, storagePath string) *Service {
	return &Service{
		repo:            repo,
		storageProvider: storageProvider,
		usage:           usage,
		maxUploadBytes:  maxUploadBytes,
		publicBaseURL:   strings.TrimRight(publicBaseURL, "/"),
		shareTTLDays:    shareTTLDays,
		storagePath:     storagePath,
		sessionStore:    NewChunkedSessionStore(),
	}
}

// ListFiles devuelve archivos visibles para el usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//   - limit: cantidad maxima.
//
// Returns:
//   - Lista de archivos o error.
func (s *Service) ListFiles(ctx context.Context, ownerID string, includeAll bool, limit int) ([]File, error) {
	files, err := s.repo.ListFiles(ctx, FileFilter{OwnerID: ownerID, IncludeAll: includeAll, Limit: limit})
	if err != nil {
		return nil, appError("media_list_failed", "No se pudo listar archivos", err)
	}
	return files, nil
}

// GetFile devuelve metadata por id y valida owner.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de archivo.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//
// Returns:
//   - Archivo encontrado o error.
func (s *Service) GetFile(ctx context.Context, id string, ownerID string, includeAll bool) (File, error) {
	file, err := s.repo.FindFileByID(ctx, strings.TrimSpace(id))
	if errors.Is(err, ErrNotFound) {
		return File{}, appError("media_not_found", "Archivo no encontrado", ErrNotFound)
	}
	if err != nil {
		return File{}, appError("media_lookup_failed", "No se pudo leer archivo", err)
	}
	if !includeAll && file.OwnerID != ownerID {
		return File{}, appError("media_forbidden", "Archivo sin permisos", ErrForbidden)
	}
	return file, nil
}

// DeleteFile borra metadata, links y archivo fisico.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de archivo.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//
// Returns:
//   - Error controlado si no pudo borrar.
func (s *Service) DeleteFile(ctx context.Context, id string, ownerID string, includeAll bool) error {
	file, err := s.GetFile(ctx, id, ownerID, includeAll)
	if err != nil {
		return err
	}
	return s.deleteFile(ctx, file, DeleteReasonUser)
}

// OpenFile abre un archivo validando permisos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de archivo.
//   - ownerID: usuario autenticado.
//   - includeAll: permite admin global.
//
// Returns:
//   - Metadata y reader o error controlado.
func (s *Service) OpenFile(ctx context.Context, id string, ownerID string, includeAll bool) (DownloadFile, error) {
	file, err := s.GetFile(ctx, id, ownerID, includeAll)
	if err != nil {
		return DownloadFile{}, err
	}
	return s.openStoredFile(ctx, file)
}

func (s *Service) openStoredFile(ctx context.Context, file File) (DownloadFile, error) {
	reader, err := s.storageProvider.Open(ctx, file.StoragePath)
	if err != nil {
		return DownloadFile{}, appError("media_open_failed", "No se pudo abrir el archivo", err)
	}
	return DownloadFile{File: file, Reader: reader}, nil
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 60
	}
	if limit > 200 {
		return 200
	}
	return limit
}
