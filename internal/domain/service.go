package domain

// Service is a barbershop service (a haircut type).
// Pure Go. No database tags, no JSON tags, no framework imports.
type Service struct {
	ID              int
	Name            string
	PriceCentavos   int
	Category        string
	IsStartingPrice bool
}

func TotalCentavos(items []Service) int {
	total := 0
	for _, s := range items {
		total += s.PriceCentavos
	}
	return total
}

// FormatPrice returns the price with an optional "+" when the price
// is a starting point rather than an exact amount.
func (s Service) FormatPrice() string {
	out := FormatCentavos(s.PriceCentavos)
	if s.IsStartingPrice {
		out += "+"
	}
	return out
}
