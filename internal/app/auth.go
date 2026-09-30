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

// Authenticate looks up the user by email and verifies the password.
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

// CreateUser creates a new user with the given role, hashing the
// password before storing it.
func (s *AuthService) CreateUser(ctx context.Context, email string, fullName string, password string, role domain.Role, dailyFloorCentavos int) (domain.User, error) {
	hash, err := s.Hasher.Hash(password)
	if err != nil {
		return domain.User{}, err
	}
	return s.Users.Create(ctx, email, fullName, hash, role, dailyFloorCentavos)
}
