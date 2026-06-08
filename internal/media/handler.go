package media

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/auth"
	"github.com/tomasacebal/go-media-api/internal/web"
)

// Handler expone endpoints Fiber para media.
//
// Args:
//   - service: servicio de media.
//   - logger: logger de la aplicacion.
//   - authMiddleware: guards de autenticacion.
//
// Returns:
//   - Handler listo para registrar rutas.
type Handler struct {
	service        *Service
	logger         *log.Logger
	authMiddleware *auth.Middleware
}

// NewHandler crea un handler HTTP de media.
//
// Args:
//   - service: servicio de media.
//   - logger: logger de la aplicacion.
//   - authMiddleware: guards de autenticacion.
//
// Returns:
//   - Handler inicializado.
func NewHandler(service *Service, logger *log.Logger, authMiddleware *auth.Middleware) *Handler {
	return &Handler{service: service, logger: logger, authMiddleware: authMiddleware}
}

// RegisterRoutes registra las rutas del modulo media.
//
// Args:
//   - app: instancia Fiber.
//
// Returns:
//   - No retorna valores.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	h.registerFileRoutes(app)
	h.registerTransferRoutes(app)
	h.registerShareRoutes(app)
	h.registerChunkedRoutes(app)
	app.Get("/s/:code", h.SharePage)
	app.Get("/s/:code/download", h.DownloadShare)
	app.Get("/api/v1/public/shares/:code", h.PublicShare)
}

// SharePage devuelve la pagina publica de descarga.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - HTML publico.
func (h *Handler) SharePage(c *fiber.Ctx) error {
	return c.Type("html").SendString(web.ShareHTML())
}

func actor(c *fiber.Ctx) (auth.SessionUser, bool) {
	return auth.CurrentUser(c)
}

func includeAll(c *fiber.Ctx, user auth.SessionUser) bool {
	return user.IsAdmin() && strings.EqualFold(c.Query("all"), "true")
}

func sendDownload(c *fiber.Ctx, download DownloadFile) error {
	file := download.File
	c.Set(fiber.HeaderContentType, file.MIMEType)
	c.Set(fiber.HeaderContentLength, strconv.FormatInt(file.SizeBytes, 10))
	c.Set(fiber.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": file.OriginalName}))
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.SendStream(download.Reader, int(file.SizeBytes))
}

func writeError(c *fiber.Ctx, err error) error {
	var quotaErr *QuotaCleanupRequiredError
	if errors.As(err, &quotaErr) {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "quota_cleanup_required",
				"message": quotaErr.Error(),
			},
			"data": quotaErr.Preview,
		})
	}

	status := fiber.StatusInternalServerError
	code := "internal_error"
	message := "Error interno"
	var appErr *AppError
	if errors.As(err, &appErr) {
		code = appErr.Code
		message = appErr.Message
		status = statusForError(appErr.Err)
	}
	return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message}})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, ErrFileTooLarge), errors.Is(err, ErrFileExceedsQuota):
		return fiber.StatusRequestEntityTooLarge
	case errors.Is(err, ErrUnsupportedType):
		return fiber.StatusUnsupportedMediaType
	case errors.Is(err, ErrFileEmpty), errors.Is(err, ErrInvalidUpload), errors.Is(err, ErrInvalidVisibility), errors.Is(err, ErrInvalidShare):
		return fiber.StatusBadRequest
	case errors.Is(err, ErrQuotaExceeded):
		return fiber.StatusConflict
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrShareExpired):
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
		"error": fiber.Map{"code": fmt.Sprintf("http_%d", status), "message": message},
	})
}
