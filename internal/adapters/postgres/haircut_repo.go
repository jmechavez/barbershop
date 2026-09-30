package postgres

import (
	"context"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HaircutRepo struct {
	DB *pgxpool.Pool
}

func NewHaircutRepo(db *pgxpool.Pool) *HaircutRepo {
	return &HaircutRepo{DB: db}
}

func (r *HaircutRepo) Create(ctx context.Context, h domain.Haircut) (domain.Haircut, error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return domain.Haircut{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO haircuts
		 (barber_id, service_id, price_centavos, discount_centavos, net_centavos, discount_reason)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, created_at`,
		h.BarberID, h.ServiceID, h.PriceCentavos,
		h.DiscountCentavos, h.NetCentavos, h.DiscountReason,
	).Scan(&h.ID, &h.CreatedAt)
	if err != nil {
		return domain.Haircut{}, err
	}

	for i := range h.Payments {
		_, err := tx.Exec(ctx,
			`INSERT INTO payments (haircut_id, method, amount_centavos)
			 VALUES ($1, $2, $3)`,
			h.ID, string(h.Payments[i].Method), h.Payments[i].AmountCentavos,
		)
		if err != nil {
			return domain.Haircut{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Haircut{}, err
	}

	return h, nil
}

func (r *HaircutRepo) ListByBarber(ctx context.Context, barberID int, since time.Time) ([]domain.Haircut, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, barber_id, service_id, price_centavos,
		        discount_centavos, net_centavos, discount_reason, created_at
		 FROM haircuts
		 WHERE barber_id = $1 AND created_at >= $2
		 ORDER BY created_at DESC`,
		barberID, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Haircut
	for rows.Next() {
		var h domain.Haircut
		if err := rows.Scan(
			&h.ID, &h.BarberID, &h.ServiceID, &h.PriceCentavos,
			&h.DiscountCentavos, &h.NetCentavos, &h.DiscountReason, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		payments, err := r.loadPayments(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Payments = payments
	}

	return out, nil
}

func (r *HaircutRepo) ListAll(ctx context.Context, since time.Time) ([]domain.Haircut, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, barber_id, service_id, price_centavos,
		        discount_centavos, net_centavos, discount_reason, created_at
		 FROM haircuts
		 WHERE created_at >= $1
		 ORDER BY created_at DESC`,
		since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Haircut
	for rows.Next() {
		var h domain.Haircut
		if err := rows.Scan(
			&h.ID, &h.BarberID, &h.ServiceID, &h.PriceCentavos,
			&h.DiscountCentavos, &h.NetCentavos, &h.DiscountReason, &h.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		payments, err := r.loadPayments(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Payments = payments
	}

	return out, nil
}

func (r *HaircutRepo) loadPayments(ctx context.Context, haircutID int) ([]domain.Payment, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT method, amount_centavos FROM payments WHERE haircut_id = $1 ORDER BY id`,
		haircutID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Payment
	for rows.Next() {
		var p domain.Payment
		var method string
		if err := rows.Scan(&method, &p.AmountCentavos); err != nil {
			return nil, err
		}
		p.Method = domain.PaymentMethod(method)
		out = append(out, p)
	}
	return out, rows.Err()
}

var _ ports.HaircutRepository = (*HaircutRepo)(nil)
