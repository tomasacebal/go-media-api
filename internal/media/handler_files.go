package media

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/auth"
)

func (h *Handler) registerFileRoutes(app *fiber.App) {
	group := app.Group("/api/v1/files")
	group.Get("/", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeRead), h.ListFiles)
	group.Post("/", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeWrite), h.UploadFiles)
	group.Get("/:id", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeRead), h.GetFile)
	group.Get("/:id/download", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeRead), h.DownloadFile)
	group.Delete("/:id", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeDelete), h.DeleteFile)
}

// UploadFiles procesa uno o mas archivos.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con metadata o preview FIFO.
func (h *Handler) UploadFiles(c *fiber.Ctx) error {
	user, _ := actor(c)
	headers, err := multipartHeaders(c)
	if err != nil {
		return writeError(c, appError("media_file_required", "El campo file es obligatorio", ErrInvalidUpload))
	}
	result, err := h.service.UploadFiles(c.UserContext(), headers, UploadInput{
		OwnerID:     user.ID,
		Visibility:  c.FormValue("visibility"),
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Category:    c.FormValue("category"),
		CreatedBy:   user.Email,
		ConfirmFIFO: parseBool(c.FormValue("confirm_fifo")),
	})
	if err != nil {
		return writeError(c, err)
	}
	h.logger.Printf("files upload owner=%s count=%d deleted_fifo=%d", user.ID, len(result.Files), len(result.DeletedFiles))
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": result})
}

// ListFiles devuelve archivos activos.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con archivos.
func (h *Handler) ListFiles(c *fiber.Ctx) error {
	user, _ := actor(c)
	files, err := h.service.ListFiles(c.UserContext(), user.ID, includeAll(c, user), c.QueryInt("limit", 60))
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": files})
}

// GetFile devuelve metadata de archivo.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con metadata.
func (h *Handler) GetFile(c *fiber.Ctx) error {
	user, _ := actor(c)
	file, err := h.service.GetFile(c.UserContext(), c.Params("id"), user.ID, includeAll(c, user))
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": file})
}

// DownloadFile descarga un archivo autenticado.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Stream del archivo o error JSON.
func (h *Handler) DownloadFile(c *fiber.Ctx) error {
	user, _ := actor(c)
	download, err := h.service.OpenFile(c.UserContext(), c.Params("id"), user.ID, includeAll(c, user))
	if err != nil {
		return writeError(c, err)
	}
	return sendDownload(c, download)
}

// DeleteFile borra archivo, links directos y cuota.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Estado 204 o error JSON.
func (h *Handler) DeleteFile(c *fiber.Ctx) error {
	user, _ := actor(c)
	if err := h.service.DeleteFile(c.UserContext(), c.Params("id"), user.ID, includeAll(c, user)); err != nil {
		return writeError(c, err)
	}
	h.logger.Printf("file delete id=%s owner=%s", c.Params("id"), user.ID)
	return c.SendStatus(fiber.StatusNoContent)
}
