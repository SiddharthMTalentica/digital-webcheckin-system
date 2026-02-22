package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"backend_webcheckin/internal/models"
	"backend_webcheckin/internal/repository"

	"github.com/google/uuid"
)

const (
	MAX_FREE_BAGGAGE = 25.0
	EXCESS_FEE_PER_KG = 10.0
)

type CheckInService struct {
	repo         *repository.Repository
	holdDuration int // in seconds
}

func NewCheckInService(repo *repository.Repository, holdDuration int) *CheckInService {
	return &CheckInService{
		repo:         repo,
		holdDuration: holdDuration,
	}
}

// LookupBooking validates PNR and last name
func (s *CheckInService) LookupBooking(ctx context.Context, pnr, lastName string) (*models.Booking, *models.CheckIn, error) {
	// Find booking
	booking, err := s.repo.FindBookingByPNR(ctx, pnr, lastName)
	if err != nil {
		return nil, nil, err
	}

	// Fetch initial seat
	initialSeat, err := s.repo.FindInitialSeatByPNR(ctx, pnr)
	if err == nil && initialSeat != "" {
		booking.InitialSeatNo = initialSeat
	}

	// Check if already checked in
	checkIn, err := s.repo.FindCheckInByBookingID(ctx, booking.ID)
	if err != nil {
		return nil, nil, err
	}

	return booking, checkIn, nil
}

// GetAvailableSeats returns seat map for a flight
func (s *CheckInService) GetAvailableSeats(ctx context.Context, flightID uuid.UUID) ([]models.Seat, error) {
	return s.repo.GetAvailableSeats(ctx, flightID)
}

// HoldSeat attempts to hold a seat for check-in
func (s *CheckInService) HoldSeat(ctx context.Context, pnr string, flightID uuid.UUID, seatNo string) (bool, int, error) {
	// Check if seat is available
	seat, err := s.repo.GetSeatByFlightAndSeatNo(ctx, flightID, seatNo)
	if err != nil {
		return false, 0, fmt.Errorf("seat not found")
	}

	// Check if seat is already confirmed (booked or checked-in)
	if seat.IsBooked || seat.CheckInID != nil {
		return false, 0, fmt.Errorf("seat already occupied")
	}

	// Try to hold the seat in Redis
	success, err := s.repo.HoldSeat(ctx, pnr, seatNo, flightID, s.holdDuration)
	if err != nil {
		return false, 0, err
	}

	if !success {
		return false, 0, fmt.Errorf("seat currently held by another passenger")
	}

	return true, s.holdDuration, nil
}

// CalculateBaggageFee calculates excess baggage fee
func (s *CheckInService) CalculateBaggageFee(weight float64) (bool, float64) {
	if weight <= MAX_FREE_BAGGAGE {
		return false, 0
	}

	excess := weight - MAX_FREE_BAGGAGE
	fee := excess * EXCESS_FEE_PER_KG
	return true, fee
}

// ProcessBaggagePayment simulates baggage fee payment
func (s *CheckInService) ProcessBaggagePayment(ctx context.Context, checkInID uuid.UUID, feeAmount float64) error {
	return s.repo.UpdateCheckInBaggageFee(ctx, checkInID, feeAmount)
}

// CompleteCheckIn completes the check-in process
func (s *CheckInService) CompleteCheckIn(ctx context.Context, booking *models.Booking, seatNo string, baggageWeight float64) (*models.CheckIn, *models.BoardingPass, error) {
	// Check if payment required
	requiresPayment, feeAmount := s.CalculateBaggageFee(baggageWeight)

	// Get seat
	seat, err := s.repo.GetSeatByFlightAndSeatNo(ctx, booking.FlightID, seatNo)
	if err != nil {
		return nil, nil, fmt.Errorf("seat not found")
	}

	// Create check-in record
	checkIn := &models.CheckIn{
		BookingID:        booking.ID,
		SeatNo:           seatNo,
		BaggageWeight:    baggageWeight,
		BaggageFeePaid:   !requiresPayment, // Auto-paid if under limit
		BaggageFeeAmount: feeAmount,
		Status:           models.CheckInStatusInProgress,
	}

	if requiresPayment {
		checkIn.Status = models.CheckInStatusWaitingPayment
		if err := s.repo.CreateCheckIn(ctx, checkIn); err != nil {
			return nil, nil, err
		}
		return checkIn, nil, fmt.Errorf("PAYMENT_REQUIRED:%v", feeAmount)
	}

	// Complete check-in
	checkIn.Status = models.CheckInStatusCompleted
	now := time.Now()
	checkIn.CompletedAt = &now

	if err := s.repo.CreateCheckIn(ctx, checkIn); err != nil {
		return nil, nil, err
	}

	// Update seat
	if err := s.repo.UpdateSeatWithCheckIn(ctx, seat.ID, checkIn.ID); err != nil {
		return nil, nil, err
	}

	// Release Redis hold
	s.repo.ReleaseSeatHold(ctx, booking.PNR, seatNo)

	// Release old seat if they changed it
	if booking.InitialSeatNo != "" && booking.InitialSeatNo != seatNo {
		oldSeat, err := s.repo.GetSeatByFlightAndSeatNo(ctx, booking.FlightID, booking.InitialSeatNo)
		if err == nil {
			s.repo.ReleaseOldSeat(ctx, oldSeat.ID)
		}
	}
	// Update booking to point to new seat if changed
	if booking.InitialSeatNo != seatNo {
		s.repo.UpdateBookingSeat(ctx, booking.PNR, seat.ID)
	}

	// Generate boarding pass
	boardingPass := s.generateBoardingPass(booking, seatNo)

	return checkIn, boardingPass, nil
}

// ResumeCheckIn resumes check-in after payment
func (s *CheckInService) ResumeCheckIn(ctx context.Context, checkInID uuid.UUID, booking *models.Booking, seatNo string) (*models.BoardingPass, error) {
	// Update status to completed
	if err := s.repo.UpdateCheckInStatus(ctx, checkInID, models.CheckInStatusCompleted); err != nil {
		return nil, err
	}

	// Get seat and update
	seat, err := s.repo.GetSeatByFlightAndSeatNo(ctx, booking.FlightID, seatNo)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateSeatWithCheckIn(ctx, seat.ID, checkInID); err != nil {
		return nil, err
	}

	// Release Redis hold
	s.repo.ReleaseSeatHold(ctx, booking.PNR, seatNo)

	// Release old seat if they changed it
	if booking.InitialSeatNo != "" && booking.InitialSeatNo != seatNo {
		oldSeat, err := s.repo.GetSeatByFlightAndSeatNo(ctx, booking.FlightID, booking.InitialSeatNo)
		if err == nil {
			s.repo.ReleaseOldSeat(ctx, oldSeat.ID)
		}
	}
	// Update booking to point to new seat if changed
	if booking.InitialSeatNo != seatNo {
		s.repo.UpdateBookingSeat(ctx, booking.PNR, seat.ID)
	}

	// Generate boarding pass
	boardingPass := s.generateBoardingPass(booking, seatNo)

	return boardingPass, nil
}

// generateBoardingPass creates boarding pass data
func (s *CheckInService) generateBoardingPass(booking *models.Booking, seatNo string) *models.BoardingPass {
	// Mock gate assignment
	gates := []string{"A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2"}
	gate := gates[rand.Intn(len(gates))]

	// Boarding time: 45 minutes before departure
	boardingTime := booking.Flight.DepartureTime.Add(-45 * time.Minute)

	return &models.BoardingPass{
		PNR:           booking.PNR,
		PassengerName: fmt.Sprintf("%s %s", booking.PassengerFirstName, booking.PassengerLastName),
		FlightCode:    booking.Flight.Code,
		Source:        booking.Flight.Source,
		Destination:   booking.Flight.Destination,
		Seat:          seatNo,
		Gate:          gate,
		BoardingTime:  boardingTime.Format("03:04 PM"),
		DepartureTime: booking.Flight.DepartureTime,
	}
}

// GeneratePNR generates a random 6-character PNR
func GeneratePNR() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude confusing chars
	pnr := make([]byte, 6)
	for i := range pnr {
		pnr[i] = chars[rand.Intn(len(chars))]
	}
	return string(pnr)
}

// SeedTestBookings creates the known test PNRs (ABC123, XYZ789) if they don't already exist
func (s *CheckInService) SeedTestBookings(ctx context.Context) error {
	// Get all flights
	var flights []models.Flight
	if err := s.repo.DB.Find(&flights).Error; err != nil {
		return err
	}

	if len(flights) == 0 {
		return fmt.Errorf("no flights found, please seed flights first")
	}

	type testBooking struct {
		PNR       string
		FlightID  string
		FirstName string
		LastName  string
		Passport  string
	}

	entries := []testBooking{
		{"ABC123", flights[0].ID.String(), "John", "Doe", "P12345678"},
		{"XYZ789", flights[1%len(flights)].ID.String(), "Jane", "Smith", "P87654321"},
	}

	for _, tb := range entries {
		// Check if already exists in webcheckin_bookings
		var count int64
		s.repo.DB.Raw("SELECT COUNT(*) FROM webcheckin_bookings WHERE pnr = ?", tb.PNR).Scan(&count)
		if count > 0 {
			continue
		}

		err := s.repo.DB.Exec(
			`INSERT INTO webcheckin_bookings (pnr, flight_id, passenger_first_name, passenger_last_name, passport_number) VALUES (?, ?, ?, ?, ?)`,
			tb.PNR, tb.FlightID, tb.FirstName, tb.LastName, tb.Passport,
		).Error
		if err != nil {
			return fmt.Errorf("failed to seed %s: %w", tb.PNR, err)
		}
	}

	return nil
}

// SeedBookings generates pre-existing bookings for testing
func (s *CheckInService) SeedBookings(ctx context.Context, count int) error {
	firstNames := []string{"John", "Jane", "Michael", "Sarah", "David", "Emily", "Robert", "Lisa", "James", "Mary"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}

	// Get all flights
	var flights []models.Flight
	if err := s.repo.DB.Find(&flights).Error; err != nil {
		return err
	}

	if len(flights) == 0 {
		return fmt.Errorf("no flights found, please seed flights first")
	}

	bookings := make([]models.Booking, 0, count)
	usedPNRs := make(map[string]bool)

	for i := 0; i < count; i++ {
		// Generate unique PNR
		var pnr string
		for {
			pnr = GeneratePNR()
			if !usedPNRs[pnr] {
				usedPNRs[pnr] = true
				break
			}
		}

		flight := flights[rand.Intn(len(flights))]

		booking := models.Booking{
			PNR:                pnr,
			FlightID:           flight.ID,
			PassengerFirstName: firstNames[rand.Intn(len(firstNames))],
			PassengerLastName:  lastNames[rand.Intn(len(lastNames))],
			PassportNumber:     fmt.Sprintf("P%08d", rand.Intn(100000000)),
			Email:              fmt.Sprintf("%s@example.com", strings.ToLower(pnr)),
			Phone:              fmt.Sprintf("+1%010d", rand.Intn(10000000000)),
		}

		bookings = append(bookings, booking)
	}

	// Create test bookings with known PNRs
	testBookings := []models.Booking{
		{
			PNR:                "ABC123",
			FlightID:           flights[0].ID,
			PassengerFirstName: "John",
			PassengerLastName:  "Doe",
			PassportNumber:     "P12345678",
			Email:              "john.doe@example.com",
			Phone:              "+1234567890",
		},
		{
			PNR:                "XYZ789",
			FlightID:           flights[1 % len(flights)].ID,
			PassengerFirstName: "Jane",
			PassengerLastName:  "Smith",
			PassportNumber:     "P87654321",
			Email:              "jane.smith@example.com",
			Phone:              "+1987654321",
		},
	}

	bookings = append(testBookings, bookings...)

	return s.repo.DB.Create(&bookings).Error
}
