package postgres

import (
	"context"
	"errors"

	"barbershop/internal/domain"
	"barbershop/internal/ports"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.UserRepository = (*UserRepo)(nil)

type UserRepo struct {
	DB *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{DB: db}
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	email = domain.NormalizeEmail(email)

	var u domain.User
	var role string
	err := r.DB.QueryRow(ctx,
		`SELECT id, email, password_hash, role FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &role)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, ports.ErrUserNotFound
	}
	if err != nil {
		return domain.User{}, err
	}

	u.Role = domain.Role(role)
	return u, nil
}
