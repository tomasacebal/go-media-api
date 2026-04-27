package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/web"
)

// Handler expone rutas de login y api keys.
//
// Args:
//   - sessions: manager de sesiones.
//   - keys: servicio de api keys.
//   - middleware: guards de auth.
//
// Returns:
//   - Handler listo para registrar rutas.
type Handler struct {
	sessions   *SessionManager
	keys       *APIKeyService
	middleware *Middleware
}

// NewHandler crea un handler de auth.
//
// Args:
//   - sessions: manager de sesiones.
//   - keys: servicio de api keys.
//   - middleware: guards de auth.
//
// Returns:
//   - Handler inicializado.
func NewHandler(sessions *SessionManager, keys *APIKeyService, middleware *Middleware) *Handler {
	return &Handler{sessions: sessions, keys: keys, middleware: middleware}
}

// RegisterRoutes registra login, logout y api keys.
//
// Args:
//   - app: instancia Fiber.
//
// Returns:
//   - No retorna valores.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/login", h.LoginPage)
	app.Post("/login", h.Login)
	app.Post("/logout", h.middleware.RequireSessionJSON, h.Logout)
	app.Get("/api/v1/api-keys", h.middleware.RequireSessionJSON, h.ListAPIKeys)
	app.Post("/api/v1/api-keys", h.middleware.RequireSessionJSON, h.CreateAPIKey)

	group := app.Group("/api/v1/api-keys", h.middleware.RequireSessionJSON)
	group.Get("/", h.ListAPIKeys)
	group.Post("/", h.CreateAPIKey)
	group.Delete("/:id", h.RevokeAPIKey)
}

// LoginPage devuelve el formulario de login.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - HTML de login.
func (h *Handler) LoginPage(c *fiber.Ctx) error {
	if h.sessions.AuthenticateRequest(c) {
		return c.Redirect("/", fiber.StatusSeeOther)
	}
	return c.Type("html").SendString(web.LoginHTML())
}

// Login valida credenciales y crea sesion.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON o redirect segun Accept.
func (h *Handler) Login(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	if strings.Contains(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&input); err == nil {
			username = input.Username
			password = input.Password
		}
	}

	if !h.sessions.ValidateCredentials(username, password) {
		return WriteAuthError(c, fiber.StatusUnauthorized, "login_failed", "Credenciales invalidas")
	}
	if err := h.sessions.SetSession(c); err != nil {
		return WriteAuthError(c, fiber.StatusInternalServerError, "session_failed", "No se pudo crear la sesion")
	}

	if acceptsHTML(c) {
		return c.Redirect("/", fiber.StatusSeeOther)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": fiber.Map{"ok": true}})
}

// Logout cierra la sesion web.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON de confirmacion.
func (h *Handler) Logout(c *fiber.Ctx) error {
	h.sessions.ClearSession(c)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": fiber.Map{"ok": true}})
}

// ListAPIKeys devuelve api keys activas.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con api keys sin secretos.
func (h *Handler) ListAPIKeys(c *fiber.Ctx) error {
	keys, err := h.keys.List(c.UserContext())
	if err != nil {
		return WriteAuthError(c, fiber.StatusInternalServerError, "api_keys_list_failed", "No se pudieron listar las api keys")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": keys})
}

// CreateAPIKey crea una api key.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con api key y secreto visible una sola vez.
func (h *Handler) CreateAPIKey(c *fiber.Ctx) error {
	var input struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}

	if strings.Contains(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		if err := c.BodyParser(&input); err != nil {
			return WriteAuthError(c, fiber.StatusBadRequest, "api_key_invalid_payload", "Payload invalido")
		}
	} else {
		input.Name = c.FormValue("name")
		input.Scopes = splitRequestedScopes(c.FormValue("scopes"))
	}

	key, err := h.keys.Create(c.UserContext(), input.Name, input.Scopes)
	if err != nil {
		if errors.Is(err, ErrInvalidScope) {
			return WriteAuthError(c, fiber.StatusBadRequest, "api_key_invalid_scopes", "Scopes invalidos")
		}
		return WriteAuthError(c, fiber.StatusInternalServerError, "api_key_create_failed", "No se pudo crear la api key")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": key})
}

// RevokeAPIKey revoca una api key.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Estado 204 o error JSON.
func (h *Handler) RevokeAPIKey(c *fiber.Ctx) error {
	if err := h.keys.Revoke(c.UserContext(), c.Params("id")); err != nil {
		if errors.Is(err, ErrAPIKeyNotFound) {
			return WriteAuthError(c, fiber.StatusNotFound, "api_key_not_found", "Api key no encontrada")
		}
		return WriteAuthError(c, fiber.StatusInternalServerError, "api_key_revoke_failed", "No se pudo revocar la api key")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func acceptsHTML(c *fiber.Ctx) bool {
	accept := strings.ToLower(c.Get(fiber.HeaderAccept))
	return accept == "" || strings.Contains(accept, "text/html")
}

func splitRequestedScopes(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	raw := strings.Split(value, ",")
	scopes := make([]string, 0, len(raw))
	for _, scope := range raw {
		scope = strings.TrimSpace(scope)
		if scope != "" {
			scopes = append(scopes, scope)
		}
	}
	return scopes
}
