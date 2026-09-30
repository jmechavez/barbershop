package app

import (
	"context"
	"errors"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

type SalaryService struct {
	Haircuts ports.HaircutRepository
	Advances ports.CashAdvanceRepository
	Users    ports.UserRepository
}

func NewSalaryService(haircuts ports.HaircutRepository, advances ports.CashAdvanceRepository, users ports.UserRepository) *SalaryService {
	return &SalaryService{Haircuts: haircuts, Advances: advances, Users: users}
}

// ForRange computes a barber's salary summary for the given date range.
// The `to` time should be midnight of the day AFTER the last day you
// want included. For example, to cover Oct 1–7 inclusive, pass
// from = Oct 1 00:00 and to = Oct 8 00:00.
func (s *SalaryService) ForRange(ctx context.Context, barberID int, from, to time.Time) (domain.SalarySummary, error) {
	// Load the barber (for the daily floor).
	barber, err := s.Users.FindByID(ctx, barberID)
	if err != nil {
		return domain.SalarySummary{}, err
	}
	if barber.Role != domain.RoleBarber {
		return domain.SalarySummary{}, errors.New("not a barber")
	}

	// Fetch the haircuts and advances in the range.
	haircuts, err := s.Haircuts.ListByBarberRange(ctx, barberID, from, to)
	if err != nil {
		return domain.SalarySummary{}, err
	}
	advances, err := s.Advances.ListByBarberRange(ctx, barberID, from, to)
	if err != nil {
		return domain.SalarySummary{}, err
	}

	// Hand everything to the domain.
	return domain.CalculateSalary(haircuts, advances, barber.DailyFloorCentavos, from, to), nil
}
