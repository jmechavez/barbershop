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

func (r *UserRepo) FindByID(ctx context.Context, id int) (domain.User, error) {
	var u domain.User
	var role string
	err := r.DB.QueryRow(ctx,
		`SELECT id, email, password_hash, role FROM users WHERE id = $1`,
		id,
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

func (r *UserRepo) Create(ctx context.Context, email string, passwordHash string, role domain.Role) (domain.User, error) {
	email = domain.NormalizeEmail(email)

	var u domain.User
	var roleStr string
	err := r.DB.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, role`,
		email, passwordHash, string(role),
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &roleStr)
	if err != nil {
		return domain.User{}, err
	}
	u.Role = domain.Role(roleStr)
	return u, nil
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, email, password_hash, role FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.User
	for rows.Next() {
		var u domain.User
		var roleStr string
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &roleStr); err != nil {
			return nil, err
		}
		u.Role = domain.Role(roleStr)
		out = append(out, u)
	}
	return out, rows.Err()
}
