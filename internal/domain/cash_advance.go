package domain

import "time"

type CashAdvance struct {
	ID             int
	BarberID       int
	AmountCentavos int
	TakenAt        time.Time
	Note           string
}

// SumAdvances returns the total amount across the given advances.
func SumAdvances(advances []CashAdvance) int {
	total := 0
	for _, a := range advances {
		total += a.AmountCentavos
	}
	return total
}
