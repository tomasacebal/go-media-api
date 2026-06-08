package media

import (
	"archive/zip"
	"context"
	"crypto/rand"
	"io"
	"math/big"
	"net/url"
	"strings"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// ZipFiles crea un stream zip para varios archivos.
//
// Args:
//   - ctx: contexto de la operacion.
//   - files: archivos a incluir.
//
// Returns:
//   - Reader del zip.
func (s *Service) ZipFiles(ctx context.Context, files []File) io.ReadCloser {
	reader, writer := io.Pipe()
	go func() {
		zipWriter := zip.NewWriter(writer)
		for _, file := range files {
			if err := s.writeZipFile(ctx, zipWriter, file); err != nil {
				_ = zipWriter.Close()
				_ = writer.CloseWithError(err)
				return
			}
		}
		if err := zipWriter.Close(); err != nil {
			_ = writer.CloseWithError(err)
			return
		}
		_ = writer.Close()
	}()
	return reader
}

func (s *Service) writeZipFile(ctx context.Context, zipWriter *zip.Writer, file File) error {
	reader, err := s.storageProvider.Open(ctx, file.StoragePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	header := &zip.FileHeader{Name: safeZipName(file)}
	header.SetModTime(file.CreatedAt)
	header.Method = zip.Deflate
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, reader)
	return err
}

func randomBase62(length int) (string, error) {
	builder := strings.Builder{}
	builder.Grow(length)
	max := big.NewInt(int64(len(base62Alphabet)))
	for i := 0; i < length; i++ {
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		builder.WriteByte(base62Alphabet[index.Int64()])
	}
	return builder.String(), nil
}

func (s *Service) shareURL(code string) string {
	return s.publicBaseURL + "/s/" + url.PathEscape(code)
}

func resolveExpiryDays(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func safeZipName(file File) string {
	name := sanitizeFilename(file.OriginalName)
	if name == "" {
		name = file.ID
	}
	return name
}
