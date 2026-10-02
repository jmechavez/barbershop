package app

import (
	"context"
	"errors"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

// ErrPaymentsDoNotMatch is returned when the sum of a haircut's payments
// does not equal its price.
var ErrPaymentsDoNotMatch = errors.New("payments do not sum to haircut price")

// ErrNoPayments is returned when a haircut has no payment rows.
var ErrNoPayments = errors.New("at least one payment is required")

type HaircutService struct {
	Haircuts ports.HaircutRepository
	Services ports.ServiceRepository
}

func NewHaircutService(haircuts ports.HaircutRepository, services ports.ServiceRepository) *HaircutService {
	return &HaircutService{Haircuts: haircuts, Services: services}
}

// Record creates a haircut given a barber, a service, and the payments
// collected for it. It enforces:
//   - at least one payment
//   - the payments sum to the service price
//   - the service exists
//
// The price is copied from the service at the moment of the cut, so
// later price changes don't rewrite history.
func (s *HaircutService) Record(
	ctx context.Context,
	barberID int,
	serviceID int,
	discountCentavos int,
	discountReason string,
	payments []domain.Payment,
	createdAt time.Time,
) (domain.Haircut, error) {

	// Fetch the service.
	services, err := s.Services.List(ctx)
	if err != nil {
		return domain.Haircut{}, err
	}
	var svc domain.Service
	found := false
	for _, x := range services {
		if x.ID == serviceID {
			svc = x
			found = true
			break
		}
	}
	if !found {
		return domain.Haircut{}, errors.New("service not found")
	}

	// Build the haircut with computed net.
	h := domain.NewHaircutAt(barberID, serviceID, svc.PriceCentavos, discountCentavos, discountReason, payments, createdAt)

	// Validate payments.
	for _, p := range payments {
		if !p.Method.Valid() {
			return domain.Haircut{}, errors.New("invalid payment method")
		}
		if p.AmountCentavos <= 0 {
			return domain.Haircut{}, errors.New("payment amount must be positive")
		}
	}

	totalPaid := 0
	for _, p := range payments {
		totalPaid += p.AmountCentavos
	}

	// A free haircut (net = 0) must have zero payments.
	// Any other haircut must have payments summing to net.
	if h.NetCentavos == 0 {
		if totalPaid != 0 {
			return domain.Haircut{}, errors.New("free haircut should have no payments")
		}
	} else {
		if len(payments) == 0 {
			return domain.Haircut{}, ErrNoPayments
		}
		if totalPaid != h.NetCentavos {
			return domain.Haircut{}, ErrPaymentsDoNotMatch
		}
	}

	return s.Haircuts.Create(ctx, h)
}

// Today returns the barber's haircuts since midnight (local time).
func (s *HaircutService) Today(ctx context.Context, barberID int) ([]domain.Haircut, error) {
	since := startOfToday()
	return s.Haircuts.ListByBarber(ctx, barberID, since)
}

// TodayAll returns all haircuts since midnight.
func (s *HaircutService) TodayAll(ctx context.Context) ([]domain.Haircut, error) {
	since := startOfToday()
	return s.Haircuts.ListAll(ctx, since)
}

func startOfToday() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Range returns haircuts across all barbers between from and to (exclusive).
func (s *HaircutService) Range(ctx context.Context, from, to time.Time) ([]domain.Haircut, error) {
	return s.Haircuts.ListAllRange(ctx, from, to)
}

// RangeForBarber returns haircuts for a specific barber between from and to.
func (s *HaircutService) RangeForBarber(ctx context.Context, barberID int, from, to time.Time) ([]domain.Haircut, error) {
	return s.Haircuts.ListByBarberRange(ctx, barberID, from, to)
}
