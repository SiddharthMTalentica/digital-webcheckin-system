package models

import (
	"time"

	"github.com/google/uuid"
)

// CheckInStatus represents the state of a check-in process
type CheckInStatus string

const (
	CheckInStatusInProgress      CheckInStatus = "IN_PROGRESS"
	CheckInStatusWaitingPayment  CheckInStatus = "WAITING_PAYMENT"
	CheckInStatusCompleted       CheckInStatus = "COMPLETED"
)

// CheckIn represents a web check-in record
type CheckIn struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BookingID        uuid.UUID      `gorm:"type:uuid;not null;unique;index:idx_booking" json:"bookingId"`
	SeatNo           string         `gorm:"type:varchar(10);not null" json:"seatNo"`
	BaggageWeight    float64        `gorm:"type:decimal(5,2);default:0" json:"baggageWeight"`
	BaggageFeePaid   bool           `gorm:"default:false" json:"baggageFeePaid"`
	BaggageFeeAmount float64        `gorm:"type:decimal(10,2);default:0" json:"baggageFeeAmount"`
	Status           CheckInStatus  `gorm:"type:varchar(50);not null;default:'IN_PROGRESS';index:idx_status" json:"status"`
	CheckedInAt      time.Time      `gorm:"default:now()" json:"checkedInAt"`
	CompletedAt      *time.Time     `json:"completedAt,omitempty"`
	
	// Relationships
	Booking          *Booking       `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
}

func (CheckIn) TableName() string {
	return "checkins"
}

// BoardingPass represents the boarding pass data
type BoardingPass struct {
	PNR           string    `json:"pnr"`
	PassengerName string    `json:"passengerName"`
	FlightCode    string    `json:"flightCode"`
	Source        string    `json:"source"`
	Destination   string    `json:"destination"`
	Seat          string    `json:"seat"`
	Gate          string    `json:"gate"`
	BoardingTime  string    `json:"boardingTime"`
	DepartureTime time.Time `json:"departureTime"`
}
