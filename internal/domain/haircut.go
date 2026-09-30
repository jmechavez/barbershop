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
	ID            int
	BarberID      int
	ServiceID     int
	PriceCentavos int
	Payments      []Payment
	CreatedAt     time.Time
}
