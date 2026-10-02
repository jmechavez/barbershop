package domain

import "time"

type PaymentMethod string

const (
	PaymentGCash    PaymentMethod = "gcash"
	PaymentMaribank PaymentMethod = "maribank"
	PaymentCash     PaymentMethod = "cash"
)

func (p PaymentMethod) Valid() bool {
	return p == PaymentGCash || p == PaymentMaribank || p == PaymentCash
}

type Payment struct {
	Method         PaymentMethod
	AmountCentavos int
}

type Haircut struct {
	ID               int
	BarberID         int
	ServiceID        int
	PriceCentavos    int
	DiscountCentavos int
	NetCentavos      int
	DiscountReason   string
	Payments         []Payment
	CreatedAt        time.Time
}

// NewHaircutAt builds a haircut with an explicit created_at time.
func NewHaircutAt(
	barberID, serviceID, priceCentavos, discountCentavos int,
	reason string,
	payments []Payment,
	createdAt time.Time,
) Haircut {
	if discountCentavos < 0 {
		discountCentavos = 0
	}
	if discountCentavos > priceCentavos {
		discountCentavos = priceCentavos
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return Haircut{
		BarberID:         barberID,
		ServiceID:        serviceID,
		PriceCentavos:    priceCentavos,
		DiscountCentavos: discountCentavos,
		NetCentavos:      priceCentavos - discountCentavos,
		DiscountReason:   reason,
		Payments:         payments,
		CreatedAt:        createdAt,
	}
}

// SumListCentavos returns the total list price of the given haircuts.
func SumListCentavos(haircuts []Haircut) int {
	total := 0
	for _, h := range haircuts {
		total += h.PriceCentavos
	}
	return total
}

// SumNetCentavos returns the total net amount of the given haircuts.
func SumNetCentavos(haircuts []Haircut) int {
	total := 0
	for _, h := range haircuts {
		total += h.NetCentavos
	}
	return total
}
