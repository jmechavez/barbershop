package app

import (
	"context"
	"errors"
	"testing"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

// fakeUsers is a test double for ports.UserRepository.
type fakeUsers struct {
	byEmail map[string]domain.User
}

func (f *fakeUsers) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return domain.User{}, ports.ErrUserNotFound
	}
	return u, nil
}

// fakeHasher is a test double for ports.PasswordHasher.
// It does NOT hash — it just compares strings directly.
// That's fine for a test double; we're testing the use case, not bcrypt.
type fakeHasher struct {
	validPasswords map[string]string // hash -> plaintext
}

func (f *fakeHasher) Hash(plain string) (string, error) {
	return plain, nil
}

func (f *fakeHasher) Verify(plain, hash string) error {
	if f.validPasswords[hash] == plain {
		return nil
	}
	return ports.ErrInvalidPassword
}

func TestAuthenticate_Success(t *testing.T) {
	users := &fakeUsers{byEmail: map[string]domain.User{
		"juan@shop.ph": {
			ID:           1,
			Email:        "juan@shop.ph",
			PasswordHash: "hashed-secret",
			Role:         domain.RoleBarber,
		},
	}}
	hasher := &fakeHasher{validPasswords: map[string]string{
		"hashed-secret": "secret",
	}}

	svc := NewAuthService(users, hasher)

	u, err := svc.Authenticate(context.Background(), "juan@shop.ph", "secret")
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if u.Role != domain.RoleBarber {
		t.Errorf("Role = %v, want %v", u.Role, domain.RoleBarber)
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	users := &fakeUsers{byEmail: map[string]domain.User{}}
	hasher := &fakeHasher{validPasswords: map[string]string{}}

	svc := NewAuthService(users, hasher)

	_, err := svc.Authenticate(context.Background(), "nobody@shop.ph", "whatever")
	if !errors.Is(err, ports.ErrUserNotFound) {
		t.Errorf("err = %v, want ErrUserNotFound", err)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	users := &fakeUsers{byEmail: map[string]domain.User{
		"juan@shop.ph": {
			ID:           1,
			Email:        "juan@shop.ph",
			PasswordHash: "hashed-secret",
			Role:         domain.RoleBarber,
		},
	}}
	hasher := &fakeHasher{validPasswords: map[string]string{
		"hashed-secret": "secret",
	}}

	svc := NewAuthService(users, hasher)

	_, err := svc.Authenticate(context.Background(), "juan@shop.ph", "wrong")
	if !errors.Is(err, ports.ErrInvalidPassword) {
		t.Errorf("err = %v, want ErrInvalidPassword", err)
	}
}
