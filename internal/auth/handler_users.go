package auth

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) registerUserRoutes(app *fiber.App) {
	group := app.Group("/api/v1/users", h.middleware.RequireAdminJSON)
	group.Get("/", h.ListUsers)
	group.Post("/", h.CreateUser)
	group.Patch("/:id", h.UpdateUser)
}

// ListUsers devuelve usuarios para admin.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con usuarios.
func (h *Handler) ListUsers(c *fiber.Ctx) error {
	users, err := h.users.List(c.UserContext())
	if err != nil {
		return WriteAuthError(c, fiber.StatusInternalServerError, "users_list_failed", "No se pudieron listar usuarios")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": users})
}

// CreateUser crea un usuario por invitacion admin.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Usuario creado o error.
func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var input struct {
		Email        string `json:"email"`
		Name         string `json:"name"`
		Password     string `json:"password"`
		Role         string `json:"role"`
		QuotaBytes   int64  `json:"quota_bytes"`
		ShareTTLDays *int   `json:"share_ttl_days"`
	}
	if err := c.BodyParser(&input); err != nil {
		return WriteAuthError(c, fiber.StatusBadRequest, "user_invalid_payload", "Payload invalido")
	}
	user, err := h.users.Create(c.UserContext(), CreateUserInput{
		Email:              input.Email,
		Name:               input.Name,
		Password:           input.Password,
		Role:               input.Role,
		QuotaBytes:         input.QuotaBytes,
		ShareTTLDays:       valueOrZero(input.ShareTTLDays),
		UseDefaultShareTTL: input.ShareTTLDays == nil,
	})
	if err != nil {
		return WriteAuthError(c, fiber.StatusBadRequest, "user_create_failed", "No se pudo crear usuario")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": user})
}

// UpdateUser actualiza un usuario por admin.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Usuario actualizado o error.
func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var input struct {
		Name         *string `json:"name"`
		Role         *string `json:"role"`
		QuotaBytes   *int64  `json:"quota_bytes"`
		ShareTTLDays *int    `json:"share_ttl_days"`
		Disabled     *bool   `json:"disabled"`
	}
	if err := c.BodyParser(&input); err != nil {
		return WriteAuthError(c, fiber.StatusBadRequest, "user_invalid_payload", "Payload invalido")
	}
	user, err := h.users.Update(c.UserContext(), c.Params("id"), UserPatch{
		Name:         input.Name,
		Role:         input.Role,
		QuotaBytes:   input.QuotaBytes,
		ShareTTLDays: input.ShareTTLDays,
		Disabled:     input.Disabled,
	})
	if err != nil {
		return WriteAuthError(c, authStatusForError(err), "user_update_failed", userUpdateMessage(err))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": user})
}

func userUpdateMessage(err error) string {
	if errors.Is(err, ErrUserNotFound) {
		return "Usuario no encontrado"
	}
	return "No se pudo actualizar usuario"
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
