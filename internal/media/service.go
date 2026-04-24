package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/tomasacebal/go-media-api/internal/storage"
)

const sniffSize = 512

var allowedMIMEsByExtension = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".pdf":  "application/pdf",
}

// Service coordina validacion, storage y metadata de media.
//
// Args:
//   - repo: repository de metadata.
//   - storageProvider: provider de archivos.
//   - maxUploadBytes: limite maximo permitido.
//   - publicBaseURL: URL publica base.
//
// Returns:
//   - Servicio listo para handlers.
type Service struct {
	repo            Repository
	storageProvider storage.Provider
	maxUploadBytes  int64
	publicBaseURL   string
}

// NewService crea el servicio de media.
//
// Args:
//   - repo: repository de metadata.
//   - storageProvider: provider de archivos.
//   - maxUploadBytes: limite maximo permitido.
//   - publicBaseURL: URL publica base.
//
// Returns:
//   - Servicio inicializado.
func NewService(repo Repository, storageProvider storage.Provider, maxUploadBytes int64, publicBaseURL string) *Service {
	return &Service{
		repo:            repo,
		storageProvider: storageProvider,
		maxUploadBytes:  maxUploadBytes,
		publicBaseURL:   strings.TrimRight(publicBaseURL, "/"),
	}
}

// Upload valida y guarda un archivo multipart.
//
// Args:
//   - ctx: contexto de la operacion.
//   - header: archivo multipart recibido en `file`.
//   - input: campos opcionales sanitizados.
//
// Returns:
//   - Metadata creada o error controlado.
func (s *Service) Upload(ctx context.Context, header *multipart.FileHeader, input UploadInput) (File, error) {
	if header == nil {
		return File{}, appError("media_file_required", "El campo file es obligatorio", ErrInvalidUpload)
	}

	visibility, err := normalizeVisibility(input.Visibility)
	if err != nil {
		return File{}, appError("media_invalid_visibility", "visibility debe ser public o private", err)
	}

	validated, err := s.validateHeader(header)
	if err != nil {
		return File{}, err
	}

	src, err := header.Open()
	if err != nil {
		return File{}, appError("media_open_failed", "No se pudo leer el archivo", err)
	}
	defer src.Close()

	if err := validateMIME(src, validated.Extension, validated.MIMEType); err != nil {
		return File{}, err
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return File{}, appError("media_read_failed", "No se pudo preparar el archivo", err)
	}

	id := uuid.NewString()
	storedName := id + validated.Extension
	now := time.Now().UTC()
	publicURL := ""
	if visibility == VisibilityPublic {
		publicURL = s.publicDownloadURL(id)
	}

	file := File{
		ID:            id,
		OriginalName:  sanitizeFilename(header.Filename),
		StoredName:    storedName,
		MIMEType:      validated.MIMEType,
		Extension:     strings.TrimPrefix(validated.Extension, "."),
		SizeBytes:     header.Size,
		StorageDriver: StorageDriverLocal,
		StoragePath:   storedName,
		PublicURL:     publicURL,
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

	if err := s.repo.Create(ctx, file); err != nil {
		_ = s.storageProvider.Delete(ctx, file.StoragePath)
		return File{}, appError("media_metadata_failed", "No se pudo guardar la metadata", err)
	}

	return file, nil
}

// List devuelve metadata activa para galeria y clientes API.
//
// Args:
//   - ctx: contexto de la operacion.
//   - limit: cantidad maxima de resultados.
//
// Returns:
//   - Lista de metadata activa o error controlado.
func (s *Service) List(ctx context.Context, limit int) ([]File, error) {
	limit = clampLimit(limit)
	files, err := s.repo.List(ctx, limit)
	if err != nil {
		return nil, appError("media_list_failed", "No se pudo listar la metadata", err)
	}

	return files, nil
}

// Get devuelve metadata activa por id.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador publico.
//
// Returns:
//   - Metadata encontrada o error controlado.
func (s *Service) Get(ctx context.Context, id string) (File, error) {
	file, err := s.repo.FindByID(ctx, strings.TrimSpace(id))
	if errors.Is(err, ErrNotFound) {
		return File{}, appError("media_not_found", "Media no encontrada", ErrNotFound)
	}
	if err != nil {
		return File{}, appError("media_lookup_failed", "No se pudo leer la metadata", err)
	}

	return file, nil
}

// Download abre un archivo descargable si su visibilidad lo permite.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador publico.
//
// Returns:
//   - Metadata y reader o error controlado.
func (s *Service) Download(ctx context.Context, id string) (DownloadFile, error) {
	file, err := s.Get(ctx, id)
	if err != nil {
		return DownloadFile{}, err
	}
	if file.Visibility == VisibilityPrivate {
		return DownloadFile{}, appError("media_forbidden", "Media privada sin autenticacion", ErrForbidden)
	}

	reader, err := s.storageProvider.Open(ctx, file.StoragePath)
	if err != nil {
		return DownloadFile{}, appError("media_open_failed", "No se pudo abrir el archivo", err)
	}

	return DownloadFile{File: file, Reader: reader}, nil
}

// Delete borra metadata activa y archivo fisico.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador publico.
//
// Returns:
//   - Error controlado si no pudo borrar.
func (s *Service) Delete(ctx context.Context, id string) error {
	file, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDelete(ctx, file.ID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return appError("media_not_found", "Media no encontrada", ErrNotFound)
		}
		return appError("media_delete_failed", "No se pudo borrar la metadata", err)
	}

	if err := s.storageProvider.Delete(ctx, file.StoragePath); err != nil {
		return appError("media_storage_delete_failed", "No se pudo borrar el archivo", err)
	}

	return nil
}

func (s *Service) validateHeader(header *multipart.FileHeader) (validatedFile, error) {
	if header.Size <= 0 {
		return validatedFile{}, appError("media_empty_file", "El archivo esta vacio", ErrFileEmpty)
	}
	if header.Size > s.maxUploadBytes {
		return validatedFile{}, appError("media_too_large", "El archivo supera el tamaño maximo permitido", ErrFileTooLarge)
	}

	extension := strings.ToLower(filepath.Ext(header.Filename))
	expectedMIME, ok := allowedMIMEsByExtension[extension]
	if !ok {
		return validatedFile{}, appError("media_unsupported_extension", "La extension del archivo no esta permitida", ErrUnsupportedType)
	}

	return validatedFile{Extension: extension, MIMEType: expectedMIME}, nil
}

func validateMIME(file multipart.File, extension string, expectedMIME string) error {
	buffer := make([]byte, sniffSize)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return appError("media_sniff_failed", "No se pudo validar el archivo", err)
	}
	if n == 0 {
		return appError("media_empty_file", "El archivo esta vacio", ErrFileEmpty)
	}

	detectedMIME := http.DetectContentType(buffer[:n])
	if extension == ".pdf" && strings.HasPrefix(string(buffer[:n]), "%PDF-") {
		detectedMIME = "application/pdf"
	}
	if extension == ".webp" && isWebP(buffer[:n]) {
		detectedMIME = "image/webp"
	}

	if detectedMIME != expectedMIME {
		return appError(
			"media_unsupported_mime",
			fmt.Sprintf("MIME no permitido: %s", detectedMIME),
			ErrUnsupportedType,
		)
	}

	return nil
}

func normalizeVisibility(value string) (string, error) {
	visibility := strings.ToLower(strings.TrimSpace(value))
	if visibility == "" {
		return VisibilityPublic, nil
	}
	if visibility != VisibilityPublic && visibility != VisibilityPrivate {
		return "", ErrInvalidVisibility
	}
	return visibility, nil
}

func sanitizeFilename(value string) string {
	name := filepath.Base(strings.ReplaceAll(strings.ReplaceAll(value, "\\", "/"), "\x00", ""))
	name = sanitizeText(name, 255)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "archivo"
	}
	return name
}

func sanitizeText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	builder := strings.Builder{}
	builder.Grow(len(value))
	count := 0
	for _, r := range value {
		if unicode.IsControl(r) {
			continue
		}
		if count >= limit {
			break
		}
		builder.WriteRune(r)
		count++
	}

	return strings.TrimSpace(builder.String())
}

func (s *Service) publicDownloadURL(id string) string {
	return s.publicBaseURL + "/api/v1/media/" + url.PathEscape(id) + "/download"
}

func isWebP(data []byte) bool {
	return len(data) >= 12 &&
		string(data[0:4]) == "RIFF" &&
		string(data[8:12]) == "WEBP"
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 60
	}
	if limit > 100 {
		return 100
	}
	return limit
}

type validatedFile struct {
	Extension string
	MIMEType  string
}
