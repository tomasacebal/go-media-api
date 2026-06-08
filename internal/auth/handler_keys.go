package auth

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) registerAPIKeyRoutes(app *fiber.App) {
	app.Get("/api/v1/api-keys", h.middleware.RequireSessionJSON, h.ListAPIKeys)
	app.Post("/api/v1/api-keys", h.middleware.RequireSessionJSON, h.CreateAPIKey)

	group := app.Group("/api/v1/api-keys", h.middleware.RequireSessionJSON)
	group.Get("/", h.ListAPIKeys)
	group.Post("/", h.CreateAPIKey)
	group.Delete("/:id", h.RevokeAPIKey)
}

// ListAPIKeys devuelve api keys activas del usuario.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con api keys sin secretos.
func (h *Handler) ListAPIKeys(c *fiber.Ctx) error {
	user, _ := CurrentUser(c)
	keys, err := h.keys.List(c.UserContext(), user.ID)
	if err != nil {
		return WriteAuthError(c, fiber.StatusInternalServerError, "api_keys_list_failed", "No se pudieron listar las api keys")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": keys})
}

// CreateAPIKey crea una api key del usuario.
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

	user, _ := CurrentUser(c)
	key, err := h.keys.Create(c.UserContext(), user.ID, input.Name, input.Scopes)
	if err != nil {
		if errors.Is(err, ErrInvalidScope) {
			return WriteAuthError(c, fiber.StatusBadRequest, "api_key_invalid_scopes", "Scopes invalidos")
		}
		return WriteAuthError(c, fiber.StatusInternalServerError, "api_key_create_failed", "No se pudo crear la api key")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": key})
}

// RevokeAPIKey revoca una api key del usuario.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Estado 204 o error JSON.
func (h *Handler) RevokeAPIKey(c *fiber.Ctx) error {
	user, _ := CurrentUser(c)
	if err := h.keys.Revoke(c.UserContext(), user.ID, c.Params("id")); err != nil {
		if errors.Is(err, ErrAPIKeyNotFound) {
			return WriteAuthError(c, fiber.StatusNotFound, "api_key_not_found", "Api key no encontrada")
		}
		return WriteAuthError(c, fiber.StatusInternalServerError, "api_key_revoke_failed", "No se pudo revocar la api key")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
