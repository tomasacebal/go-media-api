package media

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

func detectMIME(file multipart.File) (string, error) {
	buffer := make([]byte, sniffSize)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", appError("media_sniff_failed", "No se pudo validar el archivo", err)
	}
	if n == 0 {
		return "", appError("media_empty_file", "El archivo esta vacio", ErrFileEmpty)
	}

	detectedMIME := http.DetectContentType(buffer[:n])
	if strings.HasPrefix(string(buffer[:n]), "%PDF-") {
		detectedMIME = "application/pdf"
	}
	if isWebP(buffer[:n]) {
		detectedMIME = "image/webp"
	}
	if strings.TrimSpace(detectedMIME) == "" {
		detectedMIME = "application/octet-stream"
	}
	return detectedMIME, nil
}

func normalizeVisibility(value string) (string, error) {
	visibility := strings.ToLower(strings.TrimSpace(value))
	if visibility == "" {
		return VisibilityPrivate, nil
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

func sanitizeExtension(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "." || len(value) > 32 || !strings.HasPrefix(value, ".") {
		return ""
	}
	for _, r := range strings.TrimPrefix(value, ".") {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			continue
		}
		return ""
	}
	return value
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

func newFileID() string {
	return uuid.NewString()
}

func isWebP(data []byte) bool {
	return len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}

func defaultExpiry(days int) time.Time {
	if days <= 0 {
		days = 30
	}
	return time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
}
