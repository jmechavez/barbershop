package postgres

import (
	"context"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CashAdvanceRepo struct {
	DB *pgxpool.Pool
}

func NewCashAdvanceRepo(db *pgxpool.Pool) *CashAdvanceRepo {
	return &CashAdvanceRepo{DB: db}
}

func (r *CashAdvanceRepo) Create(ctx context.Context, a domain.CashAdvance) (domain.CashAdvance, error) {
	err := r.DB.QueryRow(ctx,
		`INSERT INTO cash_advances (barber_id, amount_centavos, taken_at, note)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		a.BarberID, a.AmountCentavos, a.TakenAt, a.Note,
	).Scan(&a.ID)
	if err != nil {
		return domain.CashAdvance{}, err
	}
	return a, nil
}

func (r *CashAdvanceRepo) ListByBarber(ctx context.Context, barberID int, since time.Time) ([]domain.CashAdvance, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, barber_id, amount_centavos, taken_at, note
		 FROM cash_advances
		 WHERE barber_id = $1 AND taken_at >= $2
		 ORDER BY taken_at DESC`,
		barberID, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.CashAdvance
	for rows.Next() {
		var a domain.CashAdvance
		if err := rows.Scan(&a.ID, &a.BarberID, &a.AmountCentavos, &a.TakenAt, &a.Note); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *CashAdvanceRepo) ListByBarberRange(ctx context.Context, barberID int, from, to time.Time) ([]domain.CashAdvance, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, barber_id, amount_centavos, taken_at, note
		 FROM cash_advances
		 WHERE barber_id = $1 AND taken_at >= $2 AND taken_at < $3
		 ORDER BY taken_at ASC`,
		barberID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.CashAdvance
	for rows.Next() {
		var a domain.CashAdvance
		if err := rows.Scan(&a.ID, &a.BarberID, &a.AmountCentavos, &a.TakenAt, &a.Note); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

var _ ports.CashAdvanceRepository = (*CashAdvanceRepo)(nil)
