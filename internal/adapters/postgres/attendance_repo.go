package postgres

import (
	"context"
	"errors"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttendanceRepo struct {
	DB *pgxpool.Pool
}

func NewAttendanceRepo(db *pgxpool.Pool) *AttendanceRepo {
	return &AttendanceRepo{DB: db}
}

func (r *AttendanceRepo) Set(ctx context.Context, barberID int, day time.Time, status domain.AttendanceStatus) error {
	day = domain.StartOfDay(day)
	_, err := r.DB.Exec(ctx,
		`INSERT INTO attendance (barber_id, work_date, status)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (barber_id, work_date)
		 DO UPDATE SET status = EXCLUDED.status, marked_at = now()`,
		barberID, day, string(status),
	)
	return err
}

func (r *AttendanceRepo) GetForDay(ctx context.Context, barberID int, day time.Time) (domain.Attendance, error) {
	day = domain.StartOfDay(day)
	var a domain.Attendance
	var status string
	err := r.DB.QueryRow(ctx,
		`SELECT id, barber_id, work_date, status, marked_at
		 FROM attendance
		 WHERE barber_id = $1 AND work_date = $2`,
		barberID, day,
	).Scan(&a.ID, &a.BarberID, &a.WorkDate, &status, &a.MarkedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Attendance{}, ports.ErrAttendanceNotFound
	}
	if err != nil {
		return domain.Attendance{}, err
	}
	a.Status = domain.AttendanceStatus(status)
	return a, nil
}

func (r *AttendanceRepo) ListForBarberRange(ctx context.Context, barberID int, from, to time.Time) ([]domain.Attendance, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, barber_id, work_date, status, marked_at
		 FROM attendance
		 WHERE barber_id = $1 AND work_date >= $2 AND work_date < $3
		 ORDER BY work_date ASC`,
		barberID, domain.StartOfDay(from), domain.StartOfDay(to),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Attendance
	for rows.Next() {
		var a domain.Attendance
		var status string
		if err := rows.Scan(&a.ID, &a.BarberID, &a.WorkDate, &status, &a.MarkedAt); err != nil {
			return nil, err
		}
		a.Status = domain.AttendanceStatus(status)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AttendanceRepo) ListForDay(ctx context.Context, day time.Time) ([]domain.Attendance, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, barber_id, work_date, status, marked_at
		 FROM attendance
		 WHERE work_date = $1`,
		domain.StartOfDay(day),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Attendance
	for rows.Next() {
		var a domain.Attendance
		var status string
		if err := rows.Scan(&a.ID, &a.BarberID, &a.WorkDate, &status, &a.MarkedAt); err != nil {
			return nil, err
		}
		a.Status = domain.AttendanceStatus(status)
		out = append(out, a)
	}
	return out, rows.Err()
}

var _ ports.AttendanceRepository = (*AttendanceRepo)(nil)
