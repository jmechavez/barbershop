package domain

import "testing"

func TestFormatCentavos(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "₱0.00"},
		{15000, "₱150.00"},
		{125050, "₱1,250.50"},
		{-15000, "-₱150.00"},
	}
	for _, c := range cases {
		if got := FormatCentavos(c.in); got != c.want {
			t.Errorf("FormatCentavos(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}
