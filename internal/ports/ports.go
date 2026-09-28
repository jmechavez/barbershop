package ports

import (
	"context"

	"barbershop/internal/domain"
)

// ServiceRepository is a driven port: the domain needs a way to list services,
// but it doesn't care whether that comes from Postgres, a file, or memory.
// The interface is owned by the consumer (ports), not the implementer.
type ServiceRepository interface {
	List(ctx context.Context) ([]domain.Service, error)
}
