package models

import (
	"time"

	"github.com/google/uuid"
)

// Flight model (shared with booking service)
type Flight struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code          string    `gorm:"type:varchar(20);not null;unique" json:"code"`
	Source        string    `gorm:"type:varchar(100);not null" json:"source"`
	Destination   string    `gorm:"type:varchar(100);not null" json:"destination"`
	DepartureTime time.Time `gorm:"not null" json:"departureTime"`
	FlightType    string    `gorm:"type:varchar(10);not null" json:"flightType"` // A, B, C
	CreatedAt     time.Time `gorm:"default:now()" json:"createdAt"`
}

func (Flight) TableName() string {
	return "flights"
}

// Seat model (shared with booking service)
type Seat struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FlightID  uuid.UUID  `gorm:"type:uuid;not null" json:"flightId"`
	SeatNo    string     `gorm:"type:varchar(10);not null" json:"seat_no"`
	RowNum    int        `gorm:"column:row_num;not null" json:"row_num"`
	ColNum    string     `gorm:"column:col_num;type:varchar(5);not null" json:"col_num"`
	Category  string     `gorm:"type:varchar(20);default:'STANDARD'" json:"category"`
	IsBooked  bool       `gorm:"column:is_booked;default:false" json:"is_booked"`
	CheckInID *uuid.UUID `gorm:"type:uuid" json:"checkInId,omitempty"`
	CreatedAt time.Time  `gorm:"default:now()" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"default:now()" json:"updatedAt"`

	Flight *Flight `gorm:"foreignKey:FlightID" json:"flight,omitempty"`
}

func (Seat) TableName() string {
	return "seats"
}
