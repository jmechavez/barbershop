package ports

import (
	"context"
	"errors"

	"barbershop/internal/domain"
)

// ErrUserNotFound is returned by UserRepository.FindByEmail
// when no user matches the given email.
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
)

// ServiceRepository is a driven port: the domain needs a way to list services,
// but it doesn't care whether that comes from Postgres, a file, or memory.
// The interface is owned by the consumer (ports), not the implementer.
type ServiceRepository interface {
	List(ctx context.Context) ([]domain.Service, error)
}

// UserRepository is a driven port: the domain needs to look up users
// by email to authenticate them. It says nothing about how users are
// stored, only what the app needs to ask.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
}

// PasswordHasher is a driven port: the app needs to hash and verify
// passwords without knowing which algorithm is used.
type PasswordHasher interface {
	// Hash returns a one-way hash of the given plaintext password.
	Hash(plain string) (string, error)

	// Verify reports whether plain matches the given hash.
	// It returns nil when the password is correct, and a non-nil error
	// when it is wrong.
	Verify(plain, hash string) error
}
