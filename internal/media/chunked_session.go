package media

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	// chunkSizeBytes es el tamano objetivo de cada fragmento: 50 MB.
	chunkSizeBytes = 50 * 1024 * 1024
)

// ChunkedSession representa una sesion activa de carga fragmentada.
type ChunkedSession struct {
	ID          string
	OwnerID     string
	Filename    string
	TotalSize   int64
	ChunkSize   int64
	TotalChunks int
	TmpDir      string
	CreatedAt   time.Time
	mu          sync.Mutex
	received    map[int]struct{}
}

// MarkChunkReceived registra un fragmento como recibido de forma segura.
func (s *ChunkedSession) MarkChunkReceived(index int) {
	s.mu.Lock()
	s.received[index] = struct{}{}
	s.mu.Unlock()
}

// ReceivedCount devuelve la cantidad de fragmentos recibidos.
func (s *ChunkedSession) ReceivedCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.received)
}

// ChunkedSessionStore almacena sesiones activas en memoria.
type ChunkedSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*ChunkedSession
}

// NewChunkedSessionStore crea un store de sesiones vacio.
func NewChunkedSessionStore() *ChunkedSessionStore {
	return &ChunkedSessionStore{sessions: make(map[string]*ChunkedSession)}
}

// Create agrega una nueva sesion y devuelve su referencia.
//
// Args:
//   - id: identificador UUID de la sesion.
//   - ownerID: usuario dueno de la carga.
//   - filename: nombre original del archivo.
//   - totalSize: tamano total del archivo en bytes.
//   - tmpDir: directorio temporal donde se guardan los fragmentos.
//
// Returns:
//   - Puntero a la sesion creada.
func (st *ChunkedSessionStore) Create(id, ownerID, filename string, totalSize int64, tmpDir string) *ChunkedSession {
	totalChunks := int(math.Ceil(float64(totalSize) / float64(chunkSizeBytes)))
	if totalChunks < 1 {
		totalChunks = 1
	}
	s := &ChunkedSession{
		ID:          id,
		OwnerID:     ownerID,
		Filename:    filename,
		TotalSize:   totalSize,
		ChunkSize:   chunkSizeBytes,
		TotalChunks: totalChunks,
		TmpDir:      tmpDir,
		CreatedAt:   time.Now().UTC(),
		received:    make(map[int]struct{}),
	}
	st.mu.Lock()
	st.sessions[id] = s
	st.mu.Unlock()
	return s
}

// Get busca una sesion activa por ID.
//
// Args:
//   - id: identificador de la sesion.
//
// Returns:
//   - Sesion y true si existe, nil y false si no.
func (st *ChunkedSessionStore) Get(id string) (*ChunkedSession, bool) {
	st.mu.RLock()
	s, ok := st.sessions[id]
	st.mu.RUnlock()
	return s, ok
}

// Delete elimina una sesion del store.
//
// Args:
//   - id: identificador de la sesion a eliminar.
func (st *ChunkedSessionStore) Delete(id string) {
	st.mu.Lock()
	delete(st.sessions, id)
	st.mu.Unlock()
}

// validateUploadID verifica que un upload_id sea un UUID valido sin caracteres peligrosos.
//
// Args:
//   - id: identificador a validar.
//
// Returns:
//   - Error si el ID no es valido.
func validateUploadID(id string) error {
	if len(id) < 8 || len(id) > 64 {
		return fmt.Errorf("upload_id invalido: longitud fuera de rango")
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F') || c == '-') {
			return fmt.Errorf("upload_id contiene caracteres no permitidos")
		}
	}
	return nil
}
