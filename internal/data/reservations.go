package data

import (
	"time"

	"github.com/Teryn-Guzman/Lab-3/internal/validator"
)

type Reservation struct {
	ID        int64     `json:"id"`
	Customer  string    `json:"customer"`
	TableID   int       `json:"table_id"`
	TimeSlot  time.Time `json:"time_slot"`
	PartySize int       `json:"party_size"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"-"`
}
func ValidateReservation(v *validator.Validator, r *Reservation) {

	// Customer must be provided
	v.Check(r.Customer != "", "customer", "must be provided")

	// Customer name length limit
	v.Check(len(r.Customer) <= 100, "customer", "must not be more than 100 bytes long")

	// TableID must be positive
	v.Check(r.TableID > 0, "table_id", "must be a positive integer")

	// Party size must be at least 1
	v.Check(r.PartySize > 0, "party_size", "must be greater than 0")

	// TimeSlot must not be zero
	v.Check(!r.TimeSlot.IsZero(), "time_slot", "must be provided")

	// Status must be provided
	v.Check(r.Status != "", "status", "must be provided")
}