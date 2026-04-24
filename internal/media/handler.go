package media

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Handler expone endpoints Fiber para media.
//
// Args:
//   - service: servicio de media.
//   - logger: logger de la aplicacion.
//
// Returns:
//   - Handler listo para registrar rutas.
type Handler struct {
	service *Service
	logger  *log.Logger
}

// NewHandler crea un handler HTTP de media.
//
// Args:
//   - service: servicio de media.
//   - logger: logger de la aplicacion.
//
// Returns:
//   - Handler inicializado.
func NewHandler(service *Service, logger *log.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// RegisterRoutes registra las rutas del modulo media.
//
// Args:
//   - app: instancia Fiber.
//
// Returns:
//   - No retorna valores.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/v1/media")
	group.Post("/upload", h.Upload)
	group.Get("/:id", h.Get)
	group.Get("/:id/download", h.Download)
	group.Delete("/:id", h.Delete)
}

// Upload procesa un multipart upload.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con metadata o error.
func (h *Handler) Upload(c *fiber.Ctx) error {
	header, err := c.FormFile("file")
	if err != nil {
		return writeError(c, appError("media_file_required", "El campo file es obligatorio", ErrInvalidUpload))
	}

	file, err := h.service.Upload(c.UserContext(), header, UploadInput{
		Visibility:  c.FormValue("visibility"),
		Title:       c.FormValue("title"),
		Description: c.FormValue("description"),
		Category:    c.FormValue("category"),
	})
	if err != nil {
		return writeError(c, err)
	}

	h.logger.Printf("media upload id=%s mime=%s size=%d visibility=%s", file.ID, file.MIMEType, file.SizeBytes, file.Visibility)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": file,
	})
}

// Get devuelve metadata de media.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Respuesta JSON con metadata o error.
func (h *Handler) Get(c *fiber.Ctx) error {
	file, err := h.service.Get(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": file,
	})
}

// Download sirve el archivo asociado a una media publica.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Stream del archivo o error JSON.
func (h *Handler) Download(c *fiber.Ctx) error {
	download, err := h.service.Download(c.UserContext(), c.Params("id"))
	if err != nil {
		return writeError(c, err)
	}

	file := download.File
	c.Set(fiber.HeaderContentType, file.MIMEType)
	c.Set(fiber.HeaderContentLength, strconv.FormatInt(file.SizeBytes, 10))
	c.Set(fiber.HeaderContentDisposition, mime.FormatMediaType("inline", map[string]string{
		"filename": file.OriginalName,
	}))
	if file.Visibility == VisibilityPublic {
		c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")
	} else {
		c.Set(fiber.HeaderCacheControl, "no-store")
	}

	return c.SendStream(download.Reader, int(file.SizeBytes))
}

// Delete borra metadata y archivo fisico.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - Estado 204 o error JSON.
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Delete(c.UserContext(), id); err != nil {
		return writeError(c, err)
	}

	h.logger.Printf("media delete id=%s", id)

	return c.SendStatus(fiber.StatusNoContent)
}

func writeError(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	code := "internal_error"
	message := "Error interno"

	var appErr *AppError
	if errors.As(err, &appErr) {
		code = appErr.Code
		message = appErr.Message
		status = statusForError(appErr.Err)
	}

	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    code,
			"message": message,
		},
	})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, ErrFileTooLarge):
		return fiber.StatusRequestEntityTooLarge
	case errors.Is(err, ErrUnsupportedType):
		return fiber.StatusUnsupportedMediaType
	case errors.Is(err, ErrFileEmpty), errors.Is(err, ErrInvalidUpload), errors.Is(err, ErrInvalidVisibility):
		return fiber.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		return fiber.StatusNotFound
	case errors.Is(err, ErrForbidden):
		return fiber.StatusForbidden
	default:
		return fiber.StatusInternalServerError
	}
}

// JSONErrorHandler transforma errores Fiber en respuestas JSON.
//
// Args:
//   - c: contexto Fiber.
//   - err: error devuelto por Fiber.
//
// Returns:
//   - Respuesta JSON consistente.
func JSONErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := "Error interno"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code
		message = fiberErr.Message
	}

	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    fmt.Sprintf("http_%d", status),
			"message": message,
		},
	})
}
