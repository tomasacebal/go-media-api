package media

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/tomasacebal/go-media-api/internal/auth"
)

// registerChunkedRoutes registra los tres endpoints de carga fragmentada.
func (h *Handler) registerChunkedRoutes(app *fiber.App) {
	grp := app.Group("/api/v1/upload")
	grp.Post("/init", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeWrite), h.InitChunkedUpload)
	grp.Post("/chunk", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeWrite), h.UploadChunk)
	grp.Post("/finish", h.authMiddleware.RequireSessionOrAPIKey(auth.ScopeWrite), h.FinishChunkedUpload)
}

// initChunkedRequest es el payload de inicio de sesion de carga.
type initChunkedRequest struct {
	Filename  string `json:"filename"`
	TotalSize int64  `json:"total_size"`
}

// finishChunkedRequest es el payload de finalizacion de sesion de carga.
type finishChunkedRequest struct {
	UploadID       string   `json:"upload_id"`
	ExpectedSHA256 string   `json:"expected_sha256"`
	PeerFileIDs    []string `json:"peer_file_ids,omitempty"`
	Title          string   `json:"title,omitempty"`
	Message        string   `json:"message,omitempty"`
	ExpiresDays    int      `json:"expires_days,omitempty"`
	NeverExpires   bool     `json:"never_expires,omitempty"`
	ConfirmFIFO    bool     `json:"confirm_fifo,omitempty"`
}

// InitChunkedUpload inicia una sesion de carga fragmentada.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con upload_id, chunk_size_bytes y total_chunks.
func (h *Handler) InitChunkedUpload(c *fiber.Ctx) error {
	user, _ := actor(c)

	var req initChunkedRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, appError("invalid_json", "Payload JSON invalido", ErrInvalidUpload))
	}

	session, err := h.service.InitChunkedSession(c.UserContext(), user.ID, req.Filename, req.TotalSize)
	if err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"upload_id":        session.ID,
			"chunk_size_bytes": session.ChunkSize,
			"total_chunks":     session.TotalChunks,
		},
	})
}

// UploadChunk recibe un fragmento binario y lo guarda en disco mediante streaming.
//
// Headers obligatorios:
//   - X-Upload-ID: identificador de la sesion.
//   - X-Chunk-Index: indice 0-based del fragmento.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con estado del fragmento recibido.
func (h *Handler) UploadChunk(c *fiber.Ctx) error {
	user, _ := actor(c)

	uploadID := strings.TrimSpace(c.Get("X-Upload-ID"))
	if err := validateUploadID(uploadID); err != nil {
		return writeError(c, appError("invalid_upload_id", err.Error(), ErrInvalidUpload))
	}

	chunkIndexStr := strings.TrimSpace(c.Get("X-Chunk-Index"))
	chunkIndex, convErr := strconv.Atoi(chunkIndexStr)
	if convErr != nil || chunkIndex < 0 {
		return writeError(c, appError("invalid_chunk_index", "X-Chunk-Index debe ser un entero no negativo", ErrInvalidUpload))
	}

	session, ok := h.service.sessionStore.Get(uploadID)
	if !ok {
		return writeError(c, appError("session_not_found", "Sesion de carga no encontrada o expirada", ErrNotFound))
	}
	if session.OwnerID != user.ID {
		return writeError(c, appError("session_forbidden", "Sin permisos para esta sesion", ErrForbidden))
	}

	body := c.Body()
	if len(body) == 0 {
		return writeError(c, appError("empty_chunk", "El fragmento no puede estar vacio", ErrInvalidUpload))
	}

	// bytes.NewReader convierte el slice ya en memoria en un io.Reader;
	// io.CopyBuffer en SaveChunk lo vuelca al disco sin carga adicional.
	if err := h.service.SaveChunk(c.UserContext(), session, chunkIndex, bytes.NewReader(body)); err != nil {
		return writeError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"ok":          true,
			"chunk_index": chunkIndex,
			"received":    session.ReceivedCount(),
			"total":       session.TotalChunks,
		},
	})
}

// FinishChunkedUpload valida los fragmentos, ensambla el archivo, verifica SHA-256
// y opcionalmente crea la transferencia con su link de descarga.
//
// Args:
//   - c: contexto Fiber.
//
// Returns:
//   - JSON con file_id (sin transfer) o TransferResult completo.
func (h *Handler) FinishChunkedUpload(c *fiber.Ctx) error {
	user, _ := actor(c)

	var req finishChunkedRequest
	if err := c.BodyParser(&req); err != nil {
		return writeError(c, appError("invalid_json", "Payload JSON invalido", ErrInvalidUpload))
	}

	if err := validateUploadID(req.UploadID); err != nil {
		return writeError(c, appError("invalid_upload_id", err.Error(), ErrInvalidUpload))
	}

	session, ok := h.service.sessionStore.Get(req.UploadID)
	if !ok {
		return writeError(c, appError("session_not_found", "Sesion de carga no encontrada o expirada", ErrNotFound))
	}
	if session.OwnerID != user.ID {
		return writeError(c, appError("session_forbidden", "Sin permisos para esta sesion", ErrForbidden))
	}

	// Crear transferencia si se proveen metadatos o peer file IDs
	createTransfer := strings.TrimSpace(req.Title) != "" || len(req.PeerFileIDs) > 0 || req.NeverExpires

	expiresDays := req.ExpiresDays
	if expiresDays <= 0 && !req.NeverExpires {
		expiresDays = user.ShareTTLDays
	}

	input := TransferInput{
		UploadInput: UploadInput{
			OwnerID:     user.ID,
			Visibility:  VisibilityPrivate,
			Title:       req.Title,
			Description: req.Message,
			Category:    "transfer",
			CreatedBy:   user.Email,
			ConfirmFIFO: req.ConfirmFIFO,
		},
		Message:      req.Message,
		ExpiresDays:  expiresDays,
		NeverExpires: req.NeverExpires || expiresDays == 0,
	}

	result, err := h.service.FinishChunkedSession(
		c.UserContext(), session, req.ExpectedSHA256, req.PeerFileIDs, input, createTransfer,
	)
	if err != nil {
		return writeError(c, err)
	}

	if createTransfer {
		h.logger.Printf("chunked transfer create owner=%s peer_files=%d share=%s deleted_fifo=%d server_sha256=%s",
			user.ID, len(req.PeerFileIDs), result.Share.Code, len(result.DeletedFiles), result.ServerSHA256[:8])
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"data": fiber.Map{
				"transfer":      result.Transfer,
				"share":         result.Share,
				"files":         result.Files,
				"deleted_files": result.DeletedFiles,
				"server_sha256": result.ServerSHA256,
			},
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": fiber.Map{
			"file_id":       result.FileID,
			"server_sha256": result.ServerSHA256,
		},
	})
}
