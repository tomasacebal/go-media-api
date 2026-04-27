package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	sessionCookieName = "go_media_session"
	sessionTTL        = 12 * time.Hour
	LocalSessionUser  = "session_user"
)

// SessionManager firma y valida sesiones web.
//
// Args:
//   - username: usuario admin esperado.
//   - password: password admin esperado.
//   - secret: secreto usado para HMAC.
//
// Returns:
//   - Manager listo para emitir y validar cookies.
type SessionManager struct {
	username string
	password string
	secret   []byte
}

// NewSessionManager crea un manager de sesiones.
//
// Args:
//   - username: usuario admin esperado.
//   - password: password admin esperado.
//   - secret: secreto usado para firmar cookies.
//
// Returns:
//   - SessionManager inicializado.
func NewSessionManager(username string, password string, secret string) *SessionManager {
	return &SessionManager{
		username: strings.TrimSpace(username),
		password: password,
		secret:   []byte(secret),
	}
}

// ValidateCredentials valida usuario y password admin.
//
// Args:
//   - username: usuario recibido.
//   - password: password recibido.
//
// Returns:
//   - true si las credenciales coinciden.
func (m *SessionManager) ValidateCredentials(username string, password string) bool {
	userOK := subtle.ConstantTimeCompare([]byte(strings.TrimSpace(username)), []byte(m.username)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(m.password)) == 1
	return userOK && passOK
}

// SetSession emite una cookie firmada.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Error si no pudo firmar la sesion.
func (m *SessionManager) SetSession(c *fiber.Ctx) error {
	expiresAt := time.Now().UTC().Add(sessionTTL)
	value := m.sign(m.username, expiresAt)
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Expires:  expiresAt,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Path:     "/",
	})
	return nil
}

// ClearSession borra la cookie de sesion.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - No retorna valores.
func (m *SessionManager) ClearSession(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Path:     "/",
	})
}

// AuthenticateRequest valida la cookie y carga el usuario en locals.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - true si la sesion es valida.
func (m *SessionManager) AuthenticateRequest(c *fiber.Ctx) bool {
	username, ok := m.verify(c.Cookies(sessionCookieName))
	if !ok {
		return false
	}
	c.Locals(LocalSessionUser, username)
	return true
}

// HasSession indica si el request tiene sesion valida.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - true si existe sesion valida.
func HasSession(c *fiber.Ctx) bool {
	_, ok := c.Locals(LocalSessionUser).(string)
	return ok
}

func (m *SessionManager) sign(username string, expiresAt time.Time) string {
	expires := strconv.FormatInt(expiresAt.Unix(), 10)
	payload := base64.RawURLEncoding.EncodeToString([]byte(username)) + "." + expires
	signature := m.signature(payload)
	return payload + "." + signature
}

func (m *SessionManager) verify(value string) (string, bool) {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return "", false
	}

	payload := parts[0] + "." + parts[1]
	expected := m.signature(payload)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[2])) != 1 {
		return "", false
	}

	expiresUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Now().UTC().After(time.Unix(expiresUnix, 0)) {
		return "", false
	}

	rawUsername, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	username := string(rawUsername)
	if username != m.username {
		return "", false
	}

	return username, true
}

func (m *SessionManager) signature(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
