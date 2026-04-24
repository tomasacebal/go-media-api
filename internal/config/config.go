package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultPort            = "8080"
	defaultMaxUploadMB     = 10
	defaultStoragePath     = "media"
	defaultPublicBaseURL   = "http://localhost:8080"
	defaultSQLiteStoreName = "metadata.sqlite"
)

// Config contiene la configuracion general de la aplicacion.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Configuracion lista para inicializar Fiber y media.
type Config struct {
	Port  string
	Media MediaConfig
}

// MediaConfig contiene la configuracion del modulo de media.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Valores derivados desde variables de entorno o defaults seguros.
type MediaConfig struct {
	MaxUploadMB    int
	MaxUploadBytes int64
	StoragePath    string
	PublicBaseURL  string
	SQLitePath     string
}

// Load lee variables de entorno y devuelve una configuracion validada.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Configuracion completa o error de validacion.
func Load() (Config, error) {
	port := envString("PORT", defaultPort)
	maxUploadMB, err := envPositiveInt("MEDIA_MAX_UPLOAD_MB", defaultMaxUploadMB)
	if err != nil {
		return Config{}, err
	}

	storagePath := filepath.Clean(envString("MEDIA_STORAGE_PATH", defaultStoragePath))
	publicBaseURL := strings.TrimRight(envString("MEDIA_PUBLIC_BASE_URL", defaultPublicBaseURL), "/")
	if publicBaseURL == "" {
		return Config{}, fmt.Errorf("MEDIA_PUBLIC_BASE_URL no puede estar vacia")
	}

	return Config{
		Port: port,
		Media: MediaConfig{
			MaxUploadMB:    maxUploadMB,
			MaxUploadBytes: int64(maxUploadMB) * 1024 * 1024,
			StoragePath:    storagePath,
			PublicBaseURL:  publicBaseURL,
			SQLitePath:     filepath.Join(storagePath, defaultSQLiteStoreName),
		},
	}, nil
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envPositiveInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s debe ser un entero positivo", key)
	}

	return value, nil
}
