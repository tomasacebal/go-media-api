package media

import "context"

// Repository define persistencia de metadata de media.
//
// Args:
//   - ctx: contexto de la operacion.
//   - file: metadata a persistir.
//   - id: identificador publico.
//
// Returns:
//   - Metadata encontrada o error.
type Repository interface {
	Create(ctx context.Context, file File) error
	FindByID(ctx context.Context, id string) (File, error)
	SoftDelete(ctx context.Context, id string) error
}
