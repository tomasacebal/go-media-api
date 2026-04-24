package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// LocalProvider guarda archivos en el filesystem local.
//
// Args:
//   - rootPath: carpeta raiz del storage.
//   - publicBaseURL: base publica para futuras URLs directas.
//
// Returns:
//   - Provider local listo para guardar, abrir y borrar archivos.
type LocalProvider struct {
	rootPath      string
	publicBaseURL string
}

// NewLocalProvider crea un provider local y asegura la carpeta raiz.
//
// Args:
//   - rootPath: carpeta raiz del storage.
//   - publicBaseURL: base publica para construir URLs.
//
// Returns:
//   - Provider local o error de filesystem.
func NewLocalProvider(rootPath string, publicBaseURL string) (*LocalProvider, error) {
	cleanRoot := filepath.Clean(rootPath)
	if err := os.MkdirAll(cleanRoot, 0750); err != nil {
		return nil, fmt.Errorf("crear storage local: %w", err)
	}

	return &LocalProvider{
		rootPath:      cleanRoot,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}, nil
}

// Save guarda un archivo validado en el filesystem local.
//
// Args:
//   - ctx: contexto de la operacion.
//   - file: stream del archivo.
//   - metadata: metadata validada con path relativo.
//
// Returns:
//   - Datos del archivo guardado o error.
func (p *LocalProvider) Save(ctx context.Context, file multipart.File, metadata FileMetadata) (SavedFile, error) {
	select {
	case <-ctx.Done():
		return SavedFile{}, ctx.Err()
	default:
	}

	targetPath, err := p.resolve(metadata.StoragePath)
	if err != nil {
		return SavedFile{}, err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0750); err != nil {
		return SavedFile{}, fmt.Errorf("crear carpeta destino: %w", err)
	}

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		return SavedFile{}, fmt.Errorf("crear archivo destino: %w", err)
	}
	defer target.Close()

	written, err := io.Copy(target, file)
	if err != nil {
		return SavedFile{}, fmt.Errorf("guardar archivo: %w", err)
	}
	if written != metadata.SizeBytes {
		return SavedFile{}, fmt.Errorf("tamaño guardado invalido")
	}

	return SavedFile{StoragePath: metadata.StoragePath, SizeBytes: written}, nil
}

// Open abre un archivo por path relativo seguro.
//
// Args:
//   - ctx: contexto de la operacion.
//   - path: path relativo persistido.
//
// Returns:
//   - Reader del archivo o error.
func (p *LocalProvider) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fullPath, err := p.resolve(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("abrir archivo: %w", err)
	}

	return file, nil
}

// Delete borra un archivo por path relativo seguro.
//
// Args:
//   - ctx: contexto de la operacion.
//   - path: path relativo persistido.
//
// Returns:
//   - Error si no pudo borrar el archivo.
func (p *LocalProvider) Delete(ctx context.Context, path string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath, err := p.resolve(path)
	if err != nil {
		return err
	}

	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("borrar archivo: %w", err)
	}

	return nil
}

// PublicURL devuelve una URL directa estimada para providers publicos.
//
// Args:
//   - path: path relativo persistido.
//
// Returns:
//   - URL publica directa del provider.
func (p *LocalProvider) PublicURL(path string) string {
	cleanPath := strings.ReplaceAll(filepath.ToSlash(filepath.Clean(path)), "%", "%25")
	cleanPath = strings.TrimLeft(cleanPath, "/")
	return p.publicBaseURL + "/media/" + cleanPath
}

func (p *LocalProvider) resolve(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || cleanPath == ".." {
		return "", fmt.Errorf("path de storage invalido")
	}

	fullPath := filepath.Join(p.rootPath, cleanPath)
	rootAbs, err := filepath.Abs(p.rootPath)
	if err != nil {
		return "", fmt.Errorf("resolver storage raiz: %w", err)
	}
	fullAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("resolver path de storage: %w", err)
	}

	if fullAbs != rootAbs && !strings.HasPrefix(fullAbs, rootAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("path fuera del storage")
	}

	return fullAbs, nil
}
