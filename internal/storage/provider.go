package storage

import (
	"context"
	"io"
	"mime/multipart"
)

// FileMetadata describe el archivo que un provider debe guardar.
//
// Args:
//   - StoredName: nombre fisico seguro.
//   - StoragePath: path relativo dentro del storage.
//   - MIMEType: tipo MIME validado.
//   - SizeBytes: tamaño esperado.
//
// Returns:
//   - Metadata lista para que el provider guarde el contenido.
type FileMetadata struct {
	StoredName  string
	StoragePath string
	MIMEType    string
	SizeBytes   int64
}

// SavedFile describe el resultado de guardar un archivo.
//
// Args:
//   - StoragePath: path relativo final.
//   - SizeBytes: bytes guardados.
//
// Returns:
//   - Metadata persistible del archivo guardado.
type SavedFile struct {
	StoragePath string
	SizeBytes   int64
}

// Provider define el contrato de storage para local, S3, R2 o Supabase.
//
// Args:
//   - ctx: contexto de la operacion.
//   - file: stream del archivo.
//   - metadata: datos ya validados.
//   - path: path relativo seguro.
//
// Returns:
//   - Resultado de storage, reader, URL publica o error.
type Provider interface {
	Save(ctx context.Context, file multipart.File, metadata FileMetadata) (SavedFile, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	PublicURL(path string) string
}
