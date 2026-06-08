package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserInput = errors.New("usuario invalido")
	ErrUserDisabled     = errors.New("usuario deshabilitado")
)

// UserRepository define persistencia de usuarios.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Usuarios, cuotas y errores controlados.
type UserRepository interface {
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, user User) error
	List(ctx context.Context) ([]User, error)
	FindByID(ctx context.Context, id string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	FirstAdmin(ctx context.Context) (User, error)
	Update(ctx context.Context, user User) error
	UpdatePassword(ctx context.Context, id string, passwordHash string) error
	GetUsage(ctx context.Context, ownerID string) (int64, int64, error)
	AddStorageUsed(ctx context.Context, ownerID string, delta int64) error
}

// UserService coordina usuarios, passwords y cuotas.
//
// Args:
//   - repo: repository de usuarios.
//   - defaultQuotaBytes: cuota por defecto.
//
// Returns:
//   - Servicio listo para auth y admin.
type UserService struct {
	repo                UserRepository
	defaultQuotaBytes   int64
	defaultShareTTLDays int
}

// NewUserService crea el servicio de usuarios.
//
// Args:
//   - repo: repository de usuarios.
//   - defaultQuotaBytes: cuota default.
//
// Returns:
//   - Servicio inicializado.
func NewUserService(repo UserRepository, defaultQuotaBytes int64, defaultShareTTLDays int) *UserService {
	if defaultShareTTLDays <= 0 {
		defaultShareTTLDays = 30
	}
	return &UserService{repo: repo, defaultQuotaBytes: defaultQuotaBytes, defaultShareTTLDays: defaultShareTTLDays}
}

// EnsureBootstrapAdmin crea o devuelve el admin inicial.
//
// Args:
//   - ctx: contexto de la operacion.
//   - email: email o usuario de login admin.
//   - name: nombre visible.
//   - password: password inicial.
//   - quotaBytes: cuota del admin.
//
// Returns:
//   - Admin activo o error.
func (s *UserService) EnsureBootstrapAdmin(ctx context.Context, email string, name string, password string, quotaBytes int64, shareTTLDays int) (User, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		return User{}, err
	}
	if count > 0 {
		return s.repo.FirstAdmin(ctx)
	}
	return s.Create(ctx, CreateUserInput{
		Email:        email,
		Name:         name,
		Password:     password,
		Role:         RoleAdmin,
		QuotaBytes:   quotaBytes,
		ShareTTLDays: shareTTLDays,
	})
}

// Create crea una cuenta con password hasheado.
//
// Args:
//   - ctx: contexto de la operacion.
//   - input: datos de alta.
//
// Returns:
//   - Usuario creado o error.
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (User, error) {
	email := normalizeEmail(input.Email)
	name := sanitizeUserText(input.Name, 120)
	role := normalizeRole(input.Role)
	quota := input.QuotaBytes
	if quota <= 0 {
		quota = s.defaultQuotaBytes
	}
	shareTTLDays := input.ShareTTLDays
	if input.UseDefaultShareTTL || shareTTLDays < 0 {
		shareTTLDays = s.defaultShareTTLDays
	}
	if email == "" || len(input.Password) < 8 {
		return User{}, ErrInvalidUserInput
	}
	if name == "" {
		name = email
	}

	passwordHash, err := HashPassword(input.Password)
	if err != nil {
		return User{}, err
	}
	now := time.Now().UTC()
	user := User{
		ID:           uuid.NewString(),
		Email:        email,
		Name:         name,
		Role:         role,
		PasswordHash: passwordHash,
		QuotaBytes:   quota,
		ShareTTLDays: shareTTLDays,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return User{}, err
	}
	return user, nil
}

// Authenticate valida credenciales de usuario.
//
// Args:
//   - ctx: contexto de la operacion.
//   - email: email o usuario.
//   - password: password recibido.
//
// Returns:
//   - Usuario autenticado o error.
func (s *UserService) Authenticate(ctx context.Context, email string, password string) (User, error) {
	user, err := s.repo.FindByEmail(ctx, normalizeEmail(email))
	if err != nil {
		return User{}, err
	}
	if user.DisabledAt != nil {
		return User{}, ErrUserDisabled
	}
	if !VerifyPassword(password, user.PasswordHash) {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

// FindSessionUser devuelve un usuario activo para sesiones o keys.
//
// Args:
//   - ctx: contexto de la operacion.
//   - id: identificador de usuario.
//
// Returns:
//   - Usuario seguro o error.
func (s *UserService) FindSessionUser(ctx context.Context, id string) (SessionUser, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return SessionUser{}, err
	}
	if user.DisabledAt != nil {
		return SessionUser{}, ErrUserDisabled
	}
	return user.Session(), nil
}

// List devuelve usuarios para admin.
//
// Args:
//   - ctx: contexto de la operacion.
//
// Returns:
//   - Usuarios o error.
func (s *UserService) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}
