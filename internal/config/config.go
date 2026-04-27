package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultPort            = "8080"
	defaultMaxUploadMB     = 0
	defaultStoragePath     = "media"
	defaultPublicBaseURL   = "http://localhost:8080"
	defaultSQLiteStoreName = "metadata.sqlite"
	defaultAdminUsername   = "admin"
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
	Auth  AuthConfig
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

// AuthConfig contiene secretos y credenciales para la web.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Valores para validar login y firmar sesiones.
type AuthConfig struct {
	AdminUsername string
	AdminPassword string
	SessionSecret string
}

// Load lee variables de entorno y devuelve una configuracion validada.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Configuracion completa o error de validacion.
func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	port := envString("PORT", defaultPort)
	maxUploadMB, err := envNonNegativeInt("MEDIA_MAX_UPLOAD_MB", defaultMaxUploadMB)
	if err != nil {
		return Config{}, err
	}

	storagePath := filepath.Clean(envString("MEDIA_STORAGE_PATH", defaultStoragePath))
	publicBaseURL := strings.TrimRight(envString("MEDIA_PUBLIC_BASE_URL", defaultPublicBaseURL), "/")
	if publicBaseURL == "" {
		return Config{}, fmt.Errorf("MEDIA_PUBLIC_BASE_URL no puede estar vacia")
	}

	adminUsername := envString("ADMIN_USERNAME", defaultAdminUsername)
	adminPassword := envString("ADMIN_PASSWORD", "")
	if adminPassword == "" {
		return Config{}, fmt.Errorf("ADMIN_PASSWORD no puede estar vacia")
	}
	sessionSecret := envString("SESSION_SECRET", "")
	if len(sessionSecret) < 32 {
		return Config{}, fmt.Errorf("SESSION_SECRET debe tener al menos 32 caracteres")
	}

	maxUploadBytes := int64(0)
	if maxUploadMB > 0 {
		maxUploadBytes = int64(maxUploadMB) * 1024 * 1024
	}

	return Config{
		Port: port,
		Media: MediaConfig{
			MaxUploadMB:    maxUploadMB,
			MaxUploadBytes: maxUploadBytes,
			StoragePath:    storagePath,
			PublicBaseURL:  publicBaseURL,
			SQLitePath:     filepath.Join(storagePath, defaultSQLiteStoreName),
		},
		Auth: AuthConfig{
			AdminUsername: adminUsername,
			AdminPassword: adminPassword,
			SessionSecret: sessionSecret,
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

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("leer .env: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf(".env linea %d invalida", lineNumber)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf(".env linea %d sin clave", lineNumber)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		parsedValue, err := parseDotEnvValue(strings.TrimSpace(value), lineNumber)
		if err != nil {
			return err
		}
		if err := os.Setenv(key, parsedValue); err != nil {
			return fmt.Errorf("setear %s desde .env: %w", key, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("leer .env: %w", err)
	}

	return nil
}

func parseDotEnvValue(value string, lineNumber int) (string, error) {
	if value == "" {
		return "", nil
	}

	if strings.HasPrefix(value, `"`) {
		parsedValue, err := strconv.Unquote(value)
		if err != nil {
			return "", fmt.Errorf(".env linea %d tiene comillas invalidas", lineNumber)
		}
		return parsedValue, nil
	}
	if strings.HasPrefix(value, "'") {
		if !strings.HasSuffix(value, "'") || len(value) == 1 {
			return "", fmt.Errorf(".env linea %d tiene comillas invalidas", lineNumber)
		}
		return strings.TrimSuffix(strings.TrimPrefix(value, "'"), "'"), nil
	}

	if commentIndex := strings.Index(value, " #"); commentIndex >= 0 {
		value = value[:commentIndex]
	}
	return strings.TrimSpace(value), nil
}

func envNonNegativeInt(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s debe ser un entero mayor o igual a cero", key)
	}

	return value, nil
}
