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

const sessionCookieName = "go_media_session"

// SessionManager firma y valida sesiones web.
//
// Args:
//   - users: servicio de usuarios.
//   - secret: secreto usado para HMAC.
//   - secureCookie: activa cookies Secure.
//   - ttl: duracion de la sesion.
//
// Returns:
//   - Manager listo para emitir y validar cookies.
type SessionManager struct {
	users        *UserService
	secret       []byte
	secureCookie bool
	ttl          time.Duration
}

// NewSessionManager crea un manager de sesiones.
//
// Args:
//   - users: servicio de usuarios.
//   - secret: secreto usado para firmar cookies.
//   - secureCookie: activa cookies Secure.
//   - ttl: duracion de sesion.
//
// Returns:
//   - SessionManager inicializado.
func NewSessionManager(users *UserService, secret string, secureCookie bool, ttl time.Duration) *SessionManager {
	return &SessionManager{
		users:        users,
		secret:       []byte(secret),
		secureCookie: secureCookie,
		ttl:          ttl,
	}
}

// ValidateCredentials valida email y password.
//
// Args:
//   - c: contexto Fiber.
//   - email: email recibido.
//   - password: password recibido.
//
// Returns:
//   - Usuario de sesion, true si coincide o error.
func (m *SessionManager) ValidateCredentials(c *fiber.Ctx, email string, password string) (SessionUser, bool, error) {
	user, err := m.users.Authenticate(c.UserContext(), email, password)
	if err != nil {
		return SessionUser{}, false, err
	}
	return user.Session(), true, nil
}

// SetSession emite una cookie firmada.
//
// Args:
//   - c: contexto Fiber.
//   - userID: identificador autenticado.
//
// Returns:
//   - Error si no pudo firmar la sesion.
func (m *SessionManager) SetSession(c *fiber.Ctx, userID string) error {
	expiresAt := time.Now().UTC().Add(m.ttl)
	value := m.sign(userID, expiresAt)
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Expires:  expiresAt,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Secure:   m.secureCookie,
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
		Secure:   m.secureCookie,
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
	userID, ok := m.verify(c.Cookies(sessionCookieName))
	if !ok {
		return false
	}
	user, err := m.users.FindSessionUser(c.UserContext(), userID)
	if err != nil {
		return false
	}
	c.Locals(LocalSessionUser, user)
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
	_, ok := c.Locals(LocalSessionUser).(SessionUser)
	return ok
}

// CurrentUser devuelve el usuario autenticado.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Usuario de sesion y true si existe.
func CurrentUser(c *fiber.Ctx) (SessionUser, bool) {
	user, ok := c.Locals(LocalSessionUser).(SessionUser)
	return user, ok
}

func (m *SessionManager) sign(userID string, expiresAt time.Time) string {
	expires := strconv.FormatInt(expiresAt.Unix(), 10)
	payload := base64.RawURLEncoding.EncodeToString([]byte(userID)) + "." + expires
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
	rawUserID, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	return string(rawUserID), true
}

func (m *SessionManager) signature(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
