package ports

import (
	"context"
	"errors"

	"barbershop/internal/domain"
)

// ErrUserNotFound is returned by UserRepository.FindByEmail
// when no user matches the given email.
var ErrUserNotFound = errors.New("user not found")

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
