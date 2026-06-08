package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const LocalAPIKey = "api_key"

// Middleware agrupa guards de sesion y api key.
//
// Args:
//   - sessions: manager de sesiones web.
//   - keys: servicio de api keys.
//   - users: servicio de usuarios.
//
// Returns:
//   - Middleware listo para proteger rutas.
type Middleware struct {
	sessions *SessionManager
	keys     *APIKeyService
	users    *UserService
}

// NewMiddleware crea los guards de auth.
//
// Args:
//   - sessions: manager de sesiones web.
//   - keys: servicio de api keys.
//   - users: servicio de usuarios.
//
// Returns:
//   - Middleware inicializado.
func NewMiddleware(sessions *SessionManager, keys *APIKeyService, users *UserService) *Middleware {
	return &Middleware{sessions: sessions, keys: keys, users: users}
}

// RequireSessionPage exige sesion y redirige a login si falta.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Siguiente handler o redirect.
func (m *Middleware) RequireSessionPage(c *fiber.Ctx) error {
	if m.sessions.AuthenticateRequest(c) {
		return c.Next()
	}
	return c.Redirect("/login", fiber.StatusSeeOther)
}

// RequireSessionJSON exige sesion y devuelve JSON si falta.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Siguiente handler o error JSON.
func (m *Middleware) RequireSessionJSON(c *fiber.Ctx) error {
	if m.sessions.AuthenticateRequest(c) {
		return c.Next()
	}
	return WriteAuthError(c, fiber.StatusUnauthorized, "auth_required", "Autenticacion requerida")
}

// RequireAdminJSON exige sesion admin.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Siguiente handler o error JSON.
func (m *Middleware) RequireAdminJSON(c *fiber.Ctx) error {
	if !m.sessions.AuthenticateRequest(c) {
		return WriteAuthError(c, fiber.StatusUnauthorized, "auth_required", "Autenticacion requerida")
	}
	user, _ := CurrentUser(c)
	if !user.IsAdmin() {
		return WriteAuthError(c, fiber.StatusForbidden, "admin_required", "Admin requerido")
	}
	return c.Next()
}

// RequireAPIKey exige api key con un scope especifico.
//
// Args:
//   - scope: scope requerido.
//
// Returns:
//   - Middleware Fiber.
func (m *Middleware) RequireAPIKey(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := m.loadAPIKey(c, scope); err != nil {
			return writeAPIKeyAuthError(c, err)
		}
		if !HasAPIKey(c) {
			return WriteAuthError(c, fiber.StatusUnauthorized, "api_key_required", "Api key requerida")
		}
		return c.Next()
	}
}

// RequireSessionOrAPIKey exige sesion web o api key con scope.
//
// Args:
//   - scope: scope requerido para api keys.
//
// Returns:
//   - Middleware Fiber.
func (m *Middleware) RequireSessionOrAPIKey(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if m.sessions.AuthenticateRequest(c) {
			return c.Next()
		}
		if err := m.loadAPIKey(c, scope); err != nil {
			return writeAPIKeyAuthError(c, err)
		}
		if !HasAPIKey(c) {
			return WriteAuthError(c, fiber.StatusUnauthorized, "auth_required", "Autenticacion requerida")
		}
		return c.Next()
	}
}

// OptionalSessionOrAPIKey carga sesion o api key cuando estan presentes.
//
// Args:
//   - scope: scope requerido para aceptar api key.
//
// Returns:
//   - Middleware Fiber.
func (m *Middleware) OptionalSessionOrAPIKey(scope string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if m.sessions.AuthenticateRequest(c) {
			return c.Next()
		}
		if extractAPIKey(c) == "" {
			return c.Next()
		}
		if err := m.loadAPIKey(c, scope); err != nil {
			return writeAPIKeyAuthError(c, err)
		}
		return c.Next()
	}
}

// HasAPIKey indica si el request tiene una api key validada.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - true si hay api key validada.
func HasAPIKey(c *fiber.Ctx) bool {
	_, ok := c.Locals(LocalAPIKey).(APIKey)
	return ok
}

// WriteAuthError escribe errores JSON de auth.
//
// Args:
//   - c: contexto Fiber.
//   - status: status HTTP.
//   - code: codigo estable.
//   - message: mensaje publico.
//
// Returns:
//   - Error Fiber serializado como JSON.
func WriteAuthError(c *fiber.Ctx, status int, code string, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{"code": code, "message": message},
	})
}

func (m *Middleware) loadAPIKey(c *fiber.Ctx, scope string) error {
	raw := extractAPIKey(c)
	if raw == "" {
		return nil
	}
	key, err := m.keys.Authenticate(c.UserContext(), raw, scope)
	if err != nil {
		return err
	}
	user, err := m.users.FindSessionUser(c.UserContext(), key.OwnerID)
	if err != nil {
		return err
	}
	c.Locals(LocalAPIKey, key)
	c.Locals(LocalSessionUser, user)
	return nil
}

func writeAPIKeyAuthError(c *fiber.Ctx, err error) error {
	if errors.Is(err, ErrAPIKeyForbidden) {
		return WriteAuthError(c, fiber.StatusForbidden, "api_key_forbidden", "Api key sin permisos suficientes")
	}
	return WriteAuthError(c, fiber.StatusUnauthorized, "api_key_invalid", "Api key invalida")
}

func extractAPIKey(c *fiber.Ctx) string {
	if value := strings.TrimSpace(c.Get("X-API-Key")); value != "" {
		return value
	}
	authHeader := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}
