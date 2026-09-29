package app

import (
	"context"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

type AuthService struct {
	Users  ports.UserRepository
	Hasher ports.PasswordHasher
}

func NewAuthService(users ports.UserRepository, hasher ports.PasswordHasher) *AuthService {
	return &AuthService{Users: users, Hasher: hasher}
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	user, err := s.Users.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	if err := s.Hasher.Verify(password, user.PasswordHash); err != nil {
		return domain.User{}, err
	}

	return user, nil
}
