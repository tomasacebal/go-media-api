package media

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/auth"
)

func (h *Handler) registerTransferRoutes(app *fiber.App) {
	group := app.Group("/api/v1/transfers")
	group.Get("/", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeRead), h.ListTransfers)
	group.Post("/", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeWrite), h.CreateTransfer)
	group.Get("/:id", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeRead), h.GetTransfer)
}

// CreateTransfer crea un envio multiarchivo y link corto.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con envio y share.
func (h *Handler) CreateTransfer(c *fiber.Ctx) error {
	user, _ := actor(c)
	headers, err := multipartHeaders(c)
	if err != nil {
		return writeError(c, appError("media_file_required", "El campo file es obligatorio", ErrInvalidUpload))
	}
	expiresDays := parseInt(c.FormValue("expires_days"))
	if expiresDays <= 0 {
		expiresDays = user.ShareTTLDays
	}
	neverExpires := parseBool(c.FormValue("never_expires")) || expiresDays == 0
	result, err := h.service.CreateTransfer(c.UserContext(), headers, TransferInput{
		UploadInput: UploadInput{
			OwnerID:     user.ID,
			Visibility:  VisibilityPrivate,
			Title:       c.FormValue("title"),
			Description: c.FormValue("message"),
			Category:    "transfer",
			CreatedBy:   user.Email,
			ConfirmFIFO: parseBool(c.FormValue("confirm_fifo")),
		},
		Message:      c.FormValue("message"),
		ExpiresDays:  expiresDays,
		NeverExpires: neverExpires,
	})
	if err != nil {
		return writeError(c, err)
	}
	h.logger.Printf("transfer create owner=%s files=%d share=%s deleted_fifo=%d", user.ID, len(result.Files), result.Share.Code, len(result.DeletedFiles))
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": result})
}

// ListTransfers devuelve envios activos.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con envios.
func (h *Handler) ListTransfers(c *fiber.Ctx) error {
	user, _ := actor(c)
	transfers, err := h.service.ListTransfers(c.UserContext(), user.ID, includeAll(c, user), c.QueryInt("limit", 60))
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": transfers})
}

// GetTransfer devuelve un envio.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con envio.
func (h *Handler) GetTransfer(c *fiber.Ctx) error {
	user, _ := actor(c)
	transfer, err := h.service.GetTransfer(c.UserContext(), c.Params("id"), user.ID, includeAll(c, user))
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": transfer})
}
