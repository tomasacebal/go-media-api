package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	defaultPort            = "8080"
	defaultMaxUploadMB     = 0
	defaultStoragePath     = "media"
	defaultPublicBaseURL   = "http://localhost:8080"
	defaultSQLiteStoreName = "metadata.sqlite"
	defaultAdminUsername   = "admin"
	defaultAdminName       = "Admin"
	defaultQuotaGB         = 10
	defaultShareTTLDays    = 30
	defaultSessionTTLHours = 12
)

// Config contiene la configuracion general de la aplicacion.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Configuracion lista para inicializar Fiber y media.
type Config struct {
	Port    string
	Media   MediaConfig
	Auth    AuthConfig
	Product ProductConfig
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
	AdminUsername   string
	AdminEmail      string
	AdminName       string
	AdminPassword   string
	SessionSecret   string
	SessionTTLHours int
	CookieSecure    bool
}

// ProductConfig contiene defaults comerciales de la aplicacion.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Defaults usados para usuarios y links compartidos.
type ProductConfig struct {
	DefaultQuotaGB    int
	DefaultQuotaBytes int64
	ShareTTLDays      int
}

// Load lee variables de entorno y devuelve una configuracion validada.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Configuracion completa o error de validacion.
func Load() (Config, error) {
	if err := loadDotEnvFiles(); err != nil {
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
	adminEmail := envString("ADMIN_EMAIL", adminUsername)
	adminName := envString("ADMIN_NAME", defaultAdminName)
	adminPassword := envString("ADMIN_PASSWORD", "")
	if adminPassword == "" {
		return Config{}, fmt.Errorf("ADMIN_PASSWORD no puede estar vacia")
	}
	sessionSecret := envString("SESSION_SECRET", "")
	if len(sessionSecret) < 32 {
		return Config{}, fmt.Errorf("SESSION_SECRET debe tener al menos 32 caracteres")
	}
	sessionTTLHours, err := envPositiveInt("SESSION_TTL_HOURS", defaultSessionTTLHours)
	if err != nil {
		return Config{}, err
	}
	cookieSecure, err := envBool("COOKIE_SECURE", false)
	if err != nil {
		return Config{}, err
	}
	defaultQuotaGB, err := envPositiveInt("USER_DEFAULT_QUOTA_GB", defaultQuotaGB)
	if err != nil {
		return Config{}, err
	}
	shareTTLDays, err := envPositiveInt("SHARE_DEFAULT_TTL_DAYS", defaultShareTTLDays)
	if err != nil {
		return Config{}, err
	}

	maxUploadBytes := int64(0)
	if maxUploadMB > 0 {
		maxUploadBytes = int64(maxUploadMB) * 1024 * 1024
	}
	defaultQuotaBytes := int64(defaultQuotaGB) * 1024 * 1024 * 1024

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
			AdminUsername:   adminUsername,
			AdminEmail:      adminEmail,
			AdminName:       adminName,
			AdminPassword:   adminPassword,
			SessionSecret:   sessionSecret,
			SessionTTLHours: sessionTTLHours,
			CookieSecure:    cookieSecure,
		},
		Product: ProductConfig{
			DefaultQuotaGB:    defaultQuotaGB,
			DefaultQuotaBytes: defaultQuotaBytes,
			ShareTTLDays:      shareTTLDays,
		},
	}, nil
}
