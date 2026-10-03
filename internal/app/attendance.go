package app

import (
	"context"
	"time"

	"barbershop/internal/domain"
	"barbershop/internal/ports"
)

type AttendanceService struct {
	Repo ports.AttendanceRepository
}

func NewAttendanceService(repo ports.AttendanceRepository) *AttendanceService {
	return &AttendanceService{Repo: repo}
}

// MarkPresent records that a barber showed up on a given day.
func (s *AttendanceService) MarkPresent(ctx context.Context, barberID int, day time.Time) error {
	return s.Repo.Set(ctx, barberID, day, domain.AttendancePresent)
}

// MarkAbsent records that a barber did not show up on a given day.
func (s *AttendanceService) MarkAbsent(ctx context.Context, barberID int, day time.Time) error {
	return s.Repo.Set(ctx, barberID, day, domain.AttendanceAbsent)
}

// StatusForDay returns the status for a barber on a given day,
// or an empty string if not marked.
func (s *AttendanceService) StatusForDay(ctx context.Context, barberID int, day time.Time) (domain.AttendanceStatus, error) {
	a, err := s.Repo.GetForDay(ctx, barberID, day)
	if err != nil {
		if err == ports.ErrAttendanceNotFound {
			return "", nil
		}
		return "", err
	}
	return a.Status, nil
}

// ListForDay returns all attendance records for a specific day.
func (s *AttendanceService) ListForDay(ctx context.Context, day time.Time) ([]domain.Attendance, error) {
	return s.Repo.ListForDay(ctx, day)
}
