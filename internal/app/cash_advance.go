package app

import (
	"context"
	"errors"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

type CashAdvanceService struct {
	Advances ports.CashAdvanceRepository
}

func NewCashAdvanceService(advances ports.CashAdvanceRepository) *CashAdvanceService {
	return &CashAdvanceService{Advances: advances}
}

// Record creates a cash advance for a barber. Amount must be positive.
func (s *CashAdvanceService) Record(ctx context.Context, barberID int, amountCentavos int, note string) (domain.CashAdvance, error) {
	if amountCentavos <= 0 {
		return domain.CashAdvance{}, errors.New("amount must be positive")
	}

	a := domain.CashAdvance{
		BarberID:       barberID,
		AmountCentavos: amountCentavos,
		TakenAt:        time.Now(),
		Note:           note,
	}
	return s.Advances.Create(ctx, a)
}

// RecentForBarber returns advances for a barber in the last N days.
func (s *CashAdvanceService) RecentForBarber(ctx context.Context, barberID int, days int) ([]domain.CashAdvance, error) {
	since := time.Now().AddDate(0, 0, -days)
	return s.Advances.ListByBarber(ctx, barberID, since)
}

// ListRange returns advances for a barber between from and to (exclusive).
func (s *CashAdvanceService) ListRange(ctx context.Context, barberID int, from, to time.Time) ([]domain.CashAdvance, error) {
	return s.Advances.ListByBarberRange(ctx, barberID, from, to)
}
