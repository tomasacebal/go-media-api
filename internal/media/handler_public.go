package media

import (
	"mime"

	"github.com/gofiber/fiber/v2"
)

// PublicShare devuelve detalle publico de un link.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con archivos visibles o error.
func (h *Handler) PublicShare(c *fiber.Ctx) error {
	detail, err := h.service.ResolveShare(c.UserContext(), c.Params("code"))
	if err != nil {
		return writeError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": detail})
}

// DownloadShare descarga un link publico.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Stream de archivo o zip.
func (h *Handler) DownloadShare(c *fiber.Ctx) error {
	detail, err := h.service.OpenShare(c.UserContext(), c.Params("code"))
	if err != nil {
		return writeError(c, err)
	}
	if len(detail.Files) == 1 {
		download, err := h.service.openStoredFile(c.UserContext(), detail.Files[0])
		if err != nil {
			return writeError(c, err)
		}
		return sendDownload(c, download)
	}

	name := "envio.zip"
	if detail.Transfer != nil && detail.Transfer.Title != "" {
		name = sanitizeFilename(detail.Transfer.Title) + ".zip"
	}
	c.Set(fiber.HeaderContentType, "application/zip")
	c.Set(fiber.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.SendStream(h.service.ZipFiles(c.UserContext(), detail.Files))
}
