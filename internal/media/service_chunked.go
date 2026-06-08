package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ChunkedFinishResult es el resultado de completar una sesion de carga fragmentada.
type ChunkedFinishResult struct {
	FileID       string    `json:"file_id"`
	ServerSHA256 string    `json:"server_sha256"`
	Transfer     *Transfer `json:"transfer,omitempty"`
	Share        *Share    `json:"share,omitempty"`
	Files        []File    `json:"files,omitempty"`
	DeletedFiles []File    `json:"deleted_files,omitempty"`
}

// InitChunkedSession inicia una sesion de carga fragmentada para un archivo.
//
// Args:
//   - ctx: contexto de la operacion.
//   - ownerID: usuario dueno de la carga.
//   - filename: nombre original del archivo.
//   - totalSize: tamano total esperado en bytes.
//
// Returns:
//   - Sesion inicializada o error.
func (s *Service) InitChunkedSession(ctx context.Context, ownerID, filename string, totalSize int64) (*ChunkedSession, error) {
	if strings.TrimSpace(ownerID) == "" {
		return nil, appError("session_owner_required", "Usuario requerido", ErrInvalidUpload)
	}
	if strings.TrimSpace(filename) == "" {
		return nil, appError("session_filename_required", "Nombre de archivo requerido", ErrInvalidUpload)
	}
	if totalSize <= 0 {
		return nil, appError("session_size_invalid", "Tamano de archivo invalido", ErrInvalidUpload)
	}

	sessionID := uuid.NewString()
	tmpDir := filepath.Join(s.storagePath, "tmp", sessionID)
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		return nil, appError("session_mkdir_failed", "No se pudo crear el directorio temporal", err)
	}

	return s.sessionStore.Create(sessionID, ownerID, sanitizeFilename(filename), totalSize, tmpDir), nil
}

// SaveChunk escribe un fragmento binario en disco usando io.CopyBuffer (sin cargar en RAM).
//
// Args:
//   - ctx: contexto de la operacion.
//   - session: sesion de carga activa.
//   - chunkIndex: indice 0-based del fragmento.
//   - reader: fuente de bytes del fragmento.
//
// Returns:
//   - Error si no se pudo guardar.
func (s *Service) SaveChunk(ctx context.Context, session *ChunkedSession, chunkIndex int, reader io.Reader) error {
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return appError("chunk_index_invalid",
			fmt.Sprintf("Indice %d fuera de rango [0, %d)", chunkIndex, session.TotalChunks),
			ErrInvalidUpload)
	}

	chunkPath := filepath.Join(session.TmpDir, fmt.Sprintf("%06d.part", chunkIndex))
	f, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return appError("chunk_create_failed", "No se pudo crear el archivo de fragmento", err)
	}
	defer f.Close()

	buf := make([]byte, 4*1024*1024) // buffer de 4 MB
	if _, err := io.CopyBuffer(f, reader, buf); err != nil {
		return appError("chunk_write_failed", "Error escribiendo fragmento en disco", err)
	}

	session.MarkChunkReceived(chunkIndex)
	return nil
}

// FinishChunkedSession ensambla el archivo final, valida SHA-256 y opcionalmente crea la transferencia.
//
// Args:
//   - ctx: contexto de la operacion.
//   - session: sesion de carga a completar.
//   - expectedHash: SHA-256 hex esperado (vacio = omitir verificacion).
//   - peerFileIDs: IDs de archivos ya ensamblados en sesiones previas (para transferencias multi-archivo).
//   - input: metadatos del upload (title, message, expires, etc.).
//   - createTransfer: si true, crea la transferencia y el share.
//
// Returns:
//   - Resultado con file_id, hash y opcionalmente transfer+share.
func (s *Service) FinishChunkedSession(
	ctx context.Context,
	session *ChunkedSession,
	expectedHash string,
	peerFileIDs []string,
	input TransferInput,
	createTransfer bool,
) (ChunkedFinishResult, error) {
	// Cleanup garantizado: borrar directorio temporal y sesion sin importar el resultado.
	defer func() {
		os.RemoveAll(session.TmpDir)
		s.sessionStore.Delete(session.ID)
	}()

	// Leer y ordenar fragmentos
	entries, err := os.ReadDir(session.TmpDir)
	if err != nil {
		return ChunkedFinishResult{}, appError("finish_readdir_failed", "No se pudieron leer los fragmentos", err)
	}

	var partFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".part") {
			partFiles = append(partFiles, e.Name())
		}
	}
	sort.Strings(partFiles) // orden alfanumerico == orden de indice (000000.part, 000001.part, ...)

	if len(partFiles) == 0 {
		return ChunkedFinishResult{}, appError("no_parts_found", "No se encontraron fragmentos para ensamblar", ErrInvalidUpload)
	}
	if len(partFiles) != session.TotalChunks {
		return ChunkedFinishResult{}, appError("missing_chunks",
			fmt.Sprintf("Se esperaban %d fragmentos, se encontraron %d", session.TotalChunks, len(partFiles)),
			ErrInvalidUpload)
	}

	// Preparar ruta de salida
	fileID := newFileID()
	ext := sanitizeExtension(filepath.Ext(session.Filename))
	storedName := fileID + ext
	userDir := filepath.Join(s.storagePath, "users", session.OwnerID)
	if err := os.MkdirAll(userDir, 0750); err != nil {
		return ChunkedFinishResult{}, appError("finish_userdir_failed", "No se pudo crear el directorio de usuario", err)
	}

	outPath := filepath.Join(userDir, storedName)
	outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		return ChunkedFinishResult{}, appError("finish_create_failed", "No se pudo crear el archivo final", err)
	}

	// Escribir todos los fragmentos en orden hacia el archivo de salida,
	// calculando SHA-256 en paralelo mediante io.MultiWriter (sin carga en RAM).
	hasher := sha256.New()
	mw := io.MultiWriter(outFile, hasher)
	buf := make([]byte, 4*1024*1024)
	var totalWritten int64

	for _, partName := range partFiles {
		partPath := filepath.Join(session.TmpDir, partName)
		pf, err := os.Open(partPath)
		if err != nil {
			outFile.Close()
			os.Remove(outPath)
			return ChunkedFinishResult{}, appError("chunk_open_failed", "No se pudo abrir el fragmento: "+partName, err)
		}
		n, copyErr := io.CopyBuffer(mw, pf, buf)
		pf.Close()
		if copyErr != nil {
			outFile.Close()
			os.Remove(outPath)
			return ChunkedFinishResult{}, appError("chunk_assemble_failed", "Error al ensamblar fragmento: "+partName, copyErr)
		}
		totalWritten += n
	}
	outFile.Close()

	// Verificar integridad SHA-256
	serverHash := hex.EncodeToString(hasher.Sum(nil))
	if norm := strings.ToLower(strings.TrimSpace(expectedHash)); norm != "" && serverHash != norm {
		os.Remove(outPath)
		return ChunkedFinishResult{}, appError("hash_mismatch",
			"La integridad del archivo no pudo ser verificada (SHA-256 no coincide)", ErrInvalidUpload)
	}

	// Detectar MIME desde el archivo ensamblado
	mimeType := sniffMIMEFromPath(outPath)

	// Verificar cuota ANTES de registrar en DB
	deleted, err := s.prepareQuota(ctx, session.OwnerID, totalWritten, input.ConfirmFIFO)
	if err != nil {
		os.Remove(outPath)
		return ChunkedFinishResult{}, err
	}

	// Crear registro de archivo en base de datos
	storagePath := filepath.ToSlash(filepath.Join("users", session.OwnerID, storedName))
	now := time.Now().UTC()
	file := File{
		ID:            fileID,
		OwnerID:       session.OwnerID,
		OriginalName:  session.Filename,
		StoredName:    storedName,
		MIMEType:      mimeType,
		Extension:     strings.TrimPrefix(ext, "."),
		SizeBytes:     totalWritten,
		StorageDriver: StorageDriverLocal,
		StoragePath:   storagePath,
		Visibility:    VisibilityPrivate,
		Title:         sanitizeText(input.Title, 160),
		Description:   sanitizeText(input.Description, 600),
		Category:      sanitizeText(input.Category, 120),
		CreatedBy:     sanitizeText(input.CreatedBy, 120),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateFile(ctx, file); err != nil {
		os.Remove(outPath)
		return ChunkedFinishResult{}, appError("metadata_failed", "No se pudo registrar el archivo", err)
	}

	if err := s.usage.AddStorageUsed(ctx, session.OwnerID, totalWritten); err != nil {
		_ = s.repo.SoftDeleteFile(ctx, file.ID, DeleteReasonUser, time.Now().UTC())
		os.Remove(outPath)
		return ChunkedFinishResult{}, appError("quota_update_failed", "No se pudo actualizar la cuota", err)
	}

	result := ChunkedFinishResult{
		FileID:       file.ID,
		ServerSHA256: serverHash,
		Files:        []File{file},
		DeletedFiles: deleted,
	}

	if !createTransfer {
		return result, nil
	}

	// Crear transferencia con todos los archivos (previos + actual)
	allFileIDs := append(append([]string(nil), peerFileIDs...), file.ID)

	transfer := Transfer{
		ID:        uuid.NewString(),
		OwnerID:   session.OwnerID,
		Title:     sanitizeText(input.Title, 160),
		Message:   sanitizeText(input.Message, 600),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if transfer.Title == "" {
		transfer.Title = file.OriginalName
	}

	if err := s.repo.CreateTransfer(ctx, transfer, allFileIDs); err != nil {
		s.rollbackCreated(ctx, []File{file})
		return ChunkedFinishResult{}, appError("transfer_create_failed", "No se pudo crear la transferencia", err)
	}

	share, err := s.CreateShare(ctx, session.OwnerID, false, ShareTargetTransfer, transfer.ID,
		input.ExpiresDays, input.NeverExpires)
	if err != nil {
		s.rollbackCreated(ctx, []File{file})
		return ChunkedFinishResult{}, err
	}

	result.Transfer = &transfer
	result.Share = &share
	return result, nil
}

// sniffMIMEFromPath lee los primeros 512 bytes del archivo ensamblado para detectar su tipo MIME.
func sniffMIMEFromPath(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer f.Close()

	buf := make([]byte, sniffSize)
	n, _ := f.Read(buf)
	if n == 0 {
		return "application/octet-stream"
	}
	data := buf[:n]
	detected := http.DetectContentType(data)
	if strings.HasPrefix(string(data), "%PDF-") {
		detected = "application/pdf"
	}
	if isWebP(data) {
		detected = "image/webp"
	}
	if strings.TrimSpace(detected) == "" {
		detected = "application/octet-stream"
	}
	return detected
}
