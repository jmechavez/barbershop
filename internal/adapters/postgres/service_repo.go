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

func (r *ServiceRepo) Update(ctx context.Context, id int, name string, priceCentavos int) error {
	_, err := r.DB.Exec(ctx,
		`UPDATE services SET name = $1, price_centavos = $2 WHERE id = $3`,
		name, priceCentavos, id,
	)
	return err
}

func (r *ServiceRepo) Create(ctx context.Context, name string, priceCentavos int) (domain.Service, error) {
	var s domain.Service
	err := r.DB.QueryRow(ctx,
		`INSERT INTO services (name, price_centavos) VALUES ($1, $2)
		 RETURNING id, name, price_centavos`,
		name, priceCentavos,
	).Scan(&s.ID, &s.Name, &s.PriceCentavos)
	if err != nil {
		return domain.Service{}, err
	}
	return s, nil
}
