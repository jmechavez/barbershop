package domain

import "time"

// DailyEarning is the breakdown of a barber's earnings for one day.
type DailyEarning struct {
	Date         time.Time
	HaircutCount int
	ListTotal    int // sum of list prices for that day
	Commission   int // ListTotal × CommissionPercent / 100
	Floor        int // the barber's daily floor
	Paid         int // max(Commission, Floor)
	FloorApplied bool
}

// SalarySummary is the full breakdown for a date range.
type SalarySummary struct {
	From          time.Time
	To            time.Time
	DailyFloor    int
	Days          []DailyEarning
	GrossSalary   int // sum of Paid across days
	AdvancesTotal int
	TakeHome      int // GrossSalary − AdvancesTotal
}

// CalculateSalary computes a barber's salary between from and to.
//
// Rules:
//   - Only days with at least one haircut count.
//   - Each working day pays max(commission, dailyFloor).
//   - Commission = 50% of the sum of haircut list prices that day.
//   - Cash advances are subtracted at the end.
func CalculateSalary(
	haircuts []Haircut,
	advances []CashAdvance,
	dailyFloorCentavos int,
	from, to time.Time,
) SalarySummary {

	// Group haircuts by day (in local time).
	type dayKey struct {
		Year  int
		Month time.Month
		Day   int
	}
	byDay := map[dayKey][]Haircut{}
	for _, h := range haircuts {
		k := dayKey{h.CreatedAt.Year(), h.CreatedAt.Month(), h.CreatedAt.Day()}
		byDay[k] = append(byDay[k], h)
	}

	summary := SalarySummary{
		From:       from,
		To:         to,
		DailyFloor: dailyFloorCentavos,
	}

	// Walk every calendar day in the range.
	for d := startOfDay(from); !d.After(to); d = d.AddDate(0, 0, 1) {
		k := dayKey{d.Year(), d.Month(), d.Day()}
		dayHaircuts, ok := byDay[k]
		if !ok || len(dayHaircuts) == 0 {
			continue // no haircuts that day → not a working day
		}

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

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
