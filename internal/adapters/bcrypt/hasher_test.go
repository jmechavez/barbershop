package bcrypt

import (
	"errors"
	"testing"

	"barbershop/internal/ports"
)

func TestHashAndVerify(t *testing.T) {
	h := NewHasher()

	hash, err := h.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if hash == "correct horse battery staple" {
		t.Fatal("hash equals plaintext — you are not hashing")
	}

	if err := h.Verify("correct horse battery staple", hash); err != nil {
		t.Errorf("Verify with correct password failed: %v", err)
	}

	err = h.Verify("wrong password", hash)
	if !errors.Is(err, ports.ErrInvalidPassword) {
		t.Errorf("Verify with wrong password = %v, want ErrInvalidPassword", err)
	}
}

func TestHashIsUnique(t *testing.T) {
	// bcrypt salts internally, so hashing the same password twice
	// must produce two different hashes.
	h := NewHasher()
	a, _ := h.Hash("same password")
	b, _ := h.Hash("same password")
	if a == b {
		t.Fatal("two hashes of the same password are identical — salt is missing")
	}
}
