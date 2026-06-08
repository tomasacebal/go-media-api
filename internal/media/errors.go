package media

import "errors"

var (
	ErrInvalidUpload     = errors.New("upload invalido")
	ErrFileTooLarge      = errors.New("archivo demasiado grande")
	ErrFileEmpty         = errors.New("archivo vacio")
	ErrUnsupportedType   = errors.New("tipo de archivo no permitido")
	ErrInvalidVisibility = errors.New("visibility invalida")
	ErrNotFound          = errors.New("media no encontrada")
	ErrForbidden         = errors.New("media privada sin autenticacion")
	ErrQuotaExceeded     = errors.New("cuota excedida")
	ErrFileExceedsQuota  = errors.New("archivo supera cuota")
	ErrShareExpired      = errors.New("share vencido")
	ErrInvalidShare      = errors.New("share invalido")
)

// AppError describe errores controlados del modulo media.
//
// Args:
//   - Code: codigo estable para clientes.
//   - Message: mensaje seguro para exponer.
//   - Err: causa interna.
//
// Returns:
//   - Error serializable por handlers.
type AppError struct {
	Code    string
	Message string
	Err     error
}

// Error devuelve el mensaje publico del error.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Mensaje del error.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap devuelve la causa interna.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Error interno asociado.
func (e *AppError) Unwrap() error {
	return e.Err
}

func appError(code string, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// QuotaCleanupRequiredError indica que hace falta confirmar FIFO.
//
// Args:
//   - Preview: detalle de archivos que se borrarian.
//
// Returns:
//   - Error serializable por handlers.
type QuotaCleanupRequiredError struct {
	Preview QuotaPreview
}

// Error devuelve el mensaje publico.
//
// Args:
//   - No recibe argumentos.
//
// Returns:
//   - Mensaje del error.
func (e *QuotaCleanupRequiredError) Error() string {
	return "Se requiere confirmar limpieza de cuota"
}
