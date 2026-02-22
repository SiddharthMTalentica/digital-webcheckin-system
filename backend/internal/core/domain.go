package core

import (
	"time"

	"github.com/google/uuid"
)

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatHeld      SeatStatus = "HELD"
	SeatConfirmed SeatStatus = "CONFIRMED"
)

type Flight struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Source        string    `json:"source"`
	Destination   string    `json:"destination"`
	DepartureTime time.Time `json:"departure_time"`
	PlaneType     string    `json:"plane_type"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Seat struct {
	ID        uuid.UUID  `json:"id"`
	FlightID  uuid.UUID  `json:"flight_id"`
	SeatNo    string     `json:"seat_no"`
	RowNum    int        `json:"row_num"`
	ColNum    string     `json:"col_num"`
	Category  string     `json:"category"`
	IsBooked  bool       `json:"is_booked"`
	Status    SeatStatus `json:"status"` // Computed field
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type Booking struct {
	ID                 uuid.UUID `json:"id"`
	FlightID           uuid.UUID `json:"flight_id"`
	SeatID             *uuid.UUID `json:"seat_id"`
	PassengerFirstName string    `json:"passenger_first_name"`
	PassengerLastName  string    `json:"passenger_last_name"`
	PassengerPassport  string    `json:"passenger_passport"`
	Status             string    `json:"status"`
	BookingReference   string    `json:"booking_reference"`
	PNR                string    `json:"pnr"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
