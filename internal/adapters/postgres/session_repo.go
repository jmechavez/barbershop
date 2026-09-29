package postgres

import (
	"context"
	"errors"
	"time"

	"barbershop/internal/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepo struct {
	DB *pgxpool.Pool
}

func NewSessionRepo(db *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{DB: db}
}

func (r *SessionRepo) Create(ctx context.Context, token string, userID int, expiresAt time.Time) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	return err
}

func (r *SessionRepo) FindUserID(ctx context.Context, token string) (int, error) {
	var userID int
	err := r.DB.QueryRow(ctx,
		`SELECT user_id FROM sessions WHERE token = $1 AND expires_at > now()`,
		token,
	).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ports.ErrSessionNotFound
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *SessionRepo) Delete(ctx context.Context, token string) error {
	_, err := r.DB.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

var _ ports.SessionStore = (*SessionRepo)(nil)
