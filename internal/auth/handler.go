package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/web"
)

// Handler expone rutas de login, cuenta, usuarios y api keys.
//
// Args:
//   - sessions: manager de sesiones.
//   - users: servicio de usuarios.
//   - keys: servicio de api keys.
//   - middleware: guards de auth.
//
// Returns:
//   - Handler listo para registrar rutas.
type Handler struct {
	sessions   *SessionManager
	users      *UserService
	keys       *APIKeyService
	middleware *Middleware
}

// NewHandler crea un handler de auth.
//
// Args:
//   - sessions: manager de sesiones.
//   - users: servicio de usuarios.
//   - keys: servicio de api keys.
//   - middleware: guards de auth.
//
// Returns:
//   - Handler inicializado.
func NewHandler(sessions *SessionManager, users *UserService, keys *APIKeyService, middleware *Middleware) *Handler {
	return &Handler{sessions: sessions, users: users, keys: keys, middleware: middleware}
}

// RegisterRoutes registra rutas de auth y cuenta.
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
	app.Get("/api/v1/me", h.middleware.RequireSessionJSON, h.Me)
	app.Post("/api/v1/account/password", h.middleware.RequireSessionJSON, h.ChangePassword)
	h.registerAPIKeyRoutes(app)
	h.registerUserRoutes(app)
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
	email := c.FormValue("email")
	if email == "" {
		email = c.FormValue("username")
	}
	password := c.FormValue("password")
	if strings.Contains(c.Get(fiber.HeaderContentType), fiber.MIMEApplicationJSON) {
		var input struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&input); err == nil {
			email = input.Email
			if email == "" {
				email = input.Username
			}
			password = input.Password
		}
	}

	user, ok, err := h.sessions.ValidateCredentials(c, email, password)
	if !ok || err != nil {
		return WriteAuthError(c, fiber.StatusUnauthorized, "login_failed", "Credenciales invalidas")
	}
	if err := h.sessions.SetSession(c, user.ID); err != nil {
		return WriteAuthError(c, fiber.StatusInternalServerError, "session_failed", "No se pudo crear la sesion")
	}
	if acceptsHTML(c) {
		return c.Redirect("/", fiber.StatusSeeOther)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": user})
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

// Me devuelve la cuenta autenticada.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Usuario autenticado.
func (h *Handler) Me(c *fiber.Ctx) error {
	user, _ := CurrentUser(c)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": user})
}

// ChangePassword actualiza el password propio.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON de confirmacion o error.
func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	var input struct {
		CurrentPassword string `json:"current_password"`
		NextPassword    string `json:"next_password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return WriteAuthError(c, fiber.StatusBadRequest, "password_invalid_payload", "Payload invalido")
	}
	user, _ := CurrentUser(c)
	if err := h.users.ChangePassword(c.UserContext(), user.ID, input.CurrentPassword, input.NextPassword); err != nil {
		return WriteAuthError(c, fiber.StatusBadRequest, "password_change_failed", "No se pudo cambiar el password")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": fiber.Map{"ok": true}})
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

func authStatusForError(err error) int {
	if errors.Is(err, ErrUserNotFound) {
		return fiber.StatusNotFound
	}
	return fiber.StatusBadRequest
}
