package domain

// Service is a barbershop service (a haircut type).
// Pure Go. No database tags, no JSON tags, no framework imports.
type Service struct {
	ID            int
	Name          string
	PriceCentavos int
}

// TotalCentavos sums the prices of the given services.
// Business logic lives in the domain, not in HTTP handlers.
func TotalCentavos(items []Service) int {
	total := 0
	for _, s := range items {
		total += s.PriceCentavos
	}
	return total
}
