package domain

import "time"

// DailyEarning is the breakdown of a barber's earnings for one day.
type DailyEarning struct {
	Date         time.Time
	HaircutCount int  // how many haircuts that day
	ListTotal    int  // sum of list prices
	Commission   int  // ListTotal * CommissionPercent / 100
	Floor        int  // the barber's daily floor
	Paid         int  // max(Commission, Floor)
	FloorApplied bool // true when Paid == Floor and Commission < Floor
	Absent       bool // true when the barber was not marked present
}

// SalarySummary is the full breakdown for a date range.
type SalarySummary struct {
	From          time.Time
	To            time.Time
	DailyFloor    int
	Days          []DailyEarning
	GrossSalary   int
	AdvancesTotal int
	TakeHome      int
}

// dayKey identifies a calendar day.
type dayKey struct {
	Year  int
	Month time.Month
	Day   int
}

// CalculateSalary computes a barber's salary for a range.
//
// Rules:
//   - Days with attendance status "present" are paid max(commission, dailyFloor).
//   - Days with status "absent" are recorded with Paid = 0 and Absent = true.
//   - Days with no attendance record are skipped entirely (not in the list).
//   - Commission = CommissionPercent% of list prices that day.
//   - Cash advances are subtracted at the end.
func CalculateSalary(
	haircuts []Haircut,
	advances []CashAdvance,
	attendance []Attendance,
	dailyFloorCentavos int,
	from, to time.Time,
) SalarySummary {

	// Group haircuts by day.
	byDay := map[dayKey][]Haircut{}
	for _, h := range haircuts {
		k := dayKey{h.CreatedAt.Year(), h.CreatedAt.Month(), h.CreatedAt.Day()}
		byDay[k] = append(byDay[k], h)
	}

	// Map each attended day to its status.
	attendanceByDay := map[dayKey]AttendanceStatus{}
	for _, a := range attendance {
		k := dayKey{a.WorkDate.Year(), a.WorkDate.Month(), a.WorkDate.Day()}
		attendanceByDay[k] = a.Status
	}

	summary := SalarySummary{
		From:       from,
		To:         to,
		DailyFloor: dailyFloorCentavos,
	}

	// Walk every calendar day in the range.
	for d := startOfDayLocal(from); !d.After(to); d = d.AddDate(0, 0, 1) {
		k := dayKey{d.Year(), d.Month(), d.Day()}
		status, hasAttendance := attendanceByDay[k]

		// No record at all — skip. The barber wasn't marked for this day.
		if !hasAttendance {
			continue
		}

		// Explicitly absent — show the row with Paid = 0.
		if status == AttendanceAbsent {
			summary.Days = append(summary.Days, DailyEarning{
				Date:   d,
				Absent: true,
			})
			continue
		}

		// Present day.
		dayHaircuts := byDay[k]
		listTotal := SumListCentavos(dayHaircuts)
		commission := listTotal * CommissionPercent / 100

		paid := commission
		floorApplied := false
		if paid < dailyFloorCentavos {
			paid = dailyFloorCentavos
			floorApplied = true
		}

		summary.Days = append(summary.Days, DailyEarning{
			Date:         d,
			HaircutCount: len(dayHaircuts),
			ListTotal:    listTotal,
			Commission:   commission,
			Floor:        dailyFloorCentavos,
			Paid:         paid,
			FloorApplied: floorApplied,
		})
		summary.GrossSalary += paid
	}

	summary.AdvancesTotal = SumAdvances(advances)
	summary.TakeHome = summary.GrossSalary - summary.AdvancesTotal
	return summary
}

func startOfDayLocal(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
