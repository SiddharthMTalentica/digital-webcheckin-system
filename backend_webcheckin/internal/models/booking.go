package models

import (
	"time"

	"github.com/google/uuid"
)

// Booking represents a pre-existing flight booking (created when ticket was purchased)
type Booking struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PNR                string    `gorm:"type:varchar(6);unique;not null;index:idx_pnr" json:"pnr"`
	FlightID           uuid.UUID `gorm:"type:uuid;not null;index:idx_flight_passenger" json:"flightId"`
	PassengerFirstName string    `gorm:"column:passenger_first_name;type:varchar(100);not null" json:"passengerFirstName"`
	PassengerLastName  string    `gorm:"column:passenger_last_name;type:varchar(100);not null;index:idx_flight_passenger" json:"passengerLastName"`
	PassportNumber     string    `gorm:"column:passenger_passport;type:varchar(50)" json:"passportNumber"`
	Email              string    `gorm:"type:varchar(255)" json:"email"`
	Phone              string    `gorm:"type:varchar(20)" json:"phone"`
	CreatedAt          time.Time `gorm:"default:now()" json:"createdAt"`
	
	// Relationships
	Flight   *Flight   `gorm:"foreignKey:FlightID" json:"flight,omitempty"`
	CheckIn  *CheckIn  `gorm:"foreignKey:BookingID" json:"checkIn,omitempty"`

	InitialSeatNo string `gorm:"-" json:"initialSeatNo,omitempty"`
}

func (Booking) TableName() string {
	return "webcheckin_bookings"
}
