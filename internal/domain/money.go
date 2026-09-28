package domain

import "fmt"

// FormatCentavos renders an integer centavo amount as a Philippine peso string.
// Example: 15000 -> "₱150.00", 125050 -> "₱1,250.50"
func FormatCentavos(centavos int) string {
	sign := ""
	if centavos < 0 {
		sign = "-"
		centavos = -centavos
	}

	pesos := centavos / 100
	rem := centavos % 100

	return fmt.Sprintf("%s₱%s.%02d", sign, withThousands(pesos), rem)
}

// withThousands turns 1250 into "1,250".
func withThousands(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}

	var out []byte
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, s[i])
	}
	return string(out)
}
