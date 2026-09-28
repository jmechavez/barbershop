package domain

import "strings"

// Role is a user's role in the system. Only two exist.
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleBarber Role = "barber"
)

// Valid reports whether r is one of the two roles we accept.
func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleBarber
}

// User is a staff account (admin or barber).
// Never holds a plaintext password — only the hash.
type User struct {
	ID           int
	Email        string
	PasswordHash string
	Role         Role
}

// NormalizeEmail lowercases and trims an email so that
// "Juan@Shop.PH " and "juan@shop.ph" are the same account.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
