package domain

import "time"

// AttendanceStatus is whether a barber showed up on a given day.
type AttendanceStatus string

const (
	AttendancePresent AttendanceStatus = "present"
	AttendanceAbsent  AttendanceStatus = "absent"
)

// Attendance is a single barber's status for a single calendar day.
type Attendance struct {
	ID       int
	BarberID int
	WorkDate time.Time
	Status   AttendanceStatus
	MarkedAt time.Time
}

// StartOfDay truncates a time to midnight in its own location.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// SameDay reports whether two times fall on the same calendar day.
func SameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
