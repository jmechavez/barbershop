package postgres

import (
	"context"

	"barbershop/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ServiceRepo implements ports.ServiceRepository using pgx.
type ServiceRepo struct {
	DB *pgxpool.Pool
}

func NewServiceRepo(db *pgxpool.Pool) *ServiceRepo {
	return &ServiceRepo{DB: db}
}

func (r *ServiceRepo) List(ctx context.Context) ([]domain.Service, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, name, price_centavos FROM services ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Name, &s.PriceCentavos); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}