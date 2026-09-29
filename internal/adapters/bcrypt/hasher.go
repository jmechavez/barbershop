package bcrypt

import (
	"barbershop/internal/ports"

	"golang.org/x/crypto/bcrypt"
)

// Hasher implements ports.PasswordHasher using bcrypt.
type Hasher struct{}

func NewHasher() *Hasher {
	return &Hasher{}
}

func (h *Hasher) Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *Hasher) Verify(plain, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if err != nil {
		// Translate bcrypt's error into our sentinel.
		// We don't leak the underlying bcrypt error to callers.
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return ports.ErrInvalidPassword
		}
		return err
	}
	return nil
}

// Compile-time check that Hasher satisfies the port.
var _ ports.PasswordHasher = (*Hasher)(nil)
