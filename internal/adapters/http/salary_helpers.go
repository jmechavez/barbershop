package http

import "barbershop/internal/domain"

// countPresentDays returns how many entries in the list are NOT marked absent.
func countPresentDays(days []domain.DailyEarning) int {
	n := 0
	for _, d := range days {
		if !d.Absent {
			n++
		}
	}
	return n
}

// countAbsentDays returns how many entries in the list ARE marked absent.
func countAbsentDays(days []domain.DailyEarning) int {
	n := 0
	for _, d := range days {
		if d.Absent {
			n++
		}
	}
	return n
}
