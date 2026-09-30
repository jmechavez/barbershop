package app

import (
	"context"
	"errors"
	"testing"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

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

func (f *fakeUsers) FindByID(ctx context.Context, id int) (domain.User, error) {
	for _, u := range f.byEmail {
		if u.ID == id {
			return u, nil
		}
	}
	return domain.User{}, ports.ErrUserNotFound
}

func (f *fakeUsers) List(ctx context.Context) ([]domain.User, error) {
	var out []domain.User
	for _, u := range f.byEmail {
		out = append(out, u)
	}
	return out, nil
}

func (f *fakeUsers) Create(ctx context.Context, email string, fullName string, passwordHash string, role domain.Role, dailyFloorCentavos int) (domain.User, error) {
	u := domain.User{
		ID:                 len(f.byEmail) + 1,
		Email:              email,
		FullName:           fullName,
		PasswordHash:       passwordHash,
		Role:               role,
		DailyFloorCentavos: dailyFloorCentavos,
	}
	f.byEmail[email] = u
	return u, nil
}

type fakeHasher struct {
	validPasswords map[string]string
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
			FullName:     "Juan Dela Cruz",
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
			FullName:     "Juan Dela Cruz",
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
