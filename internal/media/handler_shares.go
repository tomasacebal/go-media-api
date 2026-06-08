package media

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/auth"
)

func (h *Handler) registerShareRoutes(app *fiber.App) {
	group := app.Group("/api/v1/shares")
	group.Get("/", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeRead), h.ListShares)
	group.Post("/", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeWrite), h.CreateShare)
	group.Delete("/:id", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeDelete), h.RevokeShare)
}

// CreateShare crea un link corto para archivo o envio.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Share creado o error.
func (h *Handler) CreateShare(c *fiber.Ctx) error {
	var input struct {
		TargetType   string `json:"target_type"`
		TargetID     string `json:"target_id"`
		ExpiresDays  int    `json:"expires_days"`
		NeverExpires bool   `json:"never_expires"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, appError("share_invalid_payload", "Payload invalido", ErrInvalidShare))
	}
	user, _ := actor(c)
	expiresDays := input.ExpiresDays
	if expiresDays <= 0 {
		expiresDays = user.ShareTTLDays
	}
	neverExpires := input.NeverExpires || expiresDays == 0
	share, err := h.service.CreateShare(c.UserContext(), user.ID, includeAll(c, user), input.TargetType, input.TargetID, expiresDays, neverExpires)
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": share})
}

// ListShares devuelve links activos.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Shares o error.
func (h *Handler) ListShares(c *fiber.Ctx) error {
	user, _ := actor(c)
	shares, err := h.service.ListShares(c.UserContext(), user.ID, includeAll(c, user), c.QueryInt("limit", 60))
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": shares})
}

// RevokeShare revoca un link.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Estado 204 o error JSON.
func (h *Handler) RevokeShare(c *fiber.Ctx) error {
	user, _ := actor(c)
	if err := h.service.RevokeShare(c.UserContext(), c.Params("id"), user.ID, includeAll(c, user)); err != nil {
		return writeError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
