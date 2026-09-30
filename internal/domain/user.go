package domain

import "strings"

// CommissionPercent is the shop-wide commission a barber earns
// on each haircut. E.g. 50 means the barber keeps 50% of the price.
const CommissionPercent = 50

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
	ID                 int
	Email              string
	FullName           string
	PasswordHash       string
	Role               Role
	DailyFloorCentavos int
}

// NormalizeEmail lowercases and trims an email so that
// "Juan@Shop.PH " and "juan@shop.ph" are the same account.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
