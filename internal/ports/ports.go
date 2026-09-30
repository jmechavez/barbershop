package ports

import (
	"context"
	"errors"
	"time"

	"barbershop/internal/domain"
)

// ErrUserNotFound is returned when no user matches a lookup.
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidPassword is returned when a password does not match the hash.
var ErrInvalidPassword = errors.New("invalid password")

// ErrSessionNotFound is returned when a session token is unknown or expired.
var ErrSessionNotFound = errors.New("session not found")

// ServiceRepository is a driven port for the services catalogue.
type ServiceRepository interface {
	List(ctx context.Context) ([]domain.Service, error)
	Create(ctx context.Context, name string, priceCentavos int) (domain.Service, error)
	Update(ctx context.Context, id int, name string, priceCentavos int) error
}

// UserRepository is a driven port for staff accounts.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id int) (domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, email string, fullName string, passwordHash string, role domain.Role, dailyFloorCentavos int) (domain.User, error)
}

// PasswordHasher is a driven port for hashing and verifying passwords.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(plain, hash string) error
}

// SessionStore is a driven port for login sessions.
type SessionStore interface {
	Create(ctx context.Context, token string, userID int, expiresAt time.Time) error
	FindUserID(ctx context.Context, token string) (int, error)
	Delete(ctx context.Context, token string) error
}

// HaircutRepository is a driven port for haircut records.
type HaircutRepository interface {
	Create(ctx context.Context, h domain.Haircut) (domain.Haircut, error)
	ListByBarber(ctx context.Context, barberID int, since time.Time) ([]domain.Haircut, error)
	ListAll(ctx context.Context, since time.Time) ([]domain.Haircut, error)
	ListByBarberRange(ctx context.Context, barberID int, from, to time.Time) ([]domain.Haircut, error)
}

type CashAdvanceRepository interface {
	Create(ctx context.Context, a domain.CashAdvance) (domain.CashAdvance, error)
	ListByBarber(ctx context.Context, barberID int, since time.Time) ([]domain.CashAdvance, error)
	ListByBarberRange(ctx context.Context, barberID int, from, to time.Time) ([]domain.CashAdvance, error)
}
