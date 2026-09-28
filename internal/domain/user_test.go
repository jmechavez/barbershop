package domain

import "testing"

func TestRoleValid(t *testing.T) {
	cases := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RoleBarber, true},
		{Role(""), false},
		{Role("superuser"), false},
		{Role("Admin"), false}, // case matters
	}
	for _, c := range cases {
		if got := c.role.Valid(); got != c.want {
			t.Errorf("Role(%q).Valid() = %v, want %v", c.role, got, c.want)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"juan@shop.ph", "juan@shop.ph"},
		{"Juan@Shop.PH", "juan@shop.ph"},
		{"  juan@shop.ph  ", "juan@shop.ph"},
		{"JUAN@SHOP.PH", "juan@shop.ph"},
	}
	for _, c := range cases {
		if got := NormalizeEmail(c.in); got != c.want {
			t.Errorf("NormalizeEmail(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
