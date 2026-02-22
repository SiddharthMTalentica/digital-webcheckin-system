package handler

import (
	"context"
	"fmt"
	"strings"

	"backend_webcheckin/internal/models"
	"backend_webcheckin/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CheckInHandler struct {
	service *service.CheckInService
}

func NewCheckInHandler(service *service.CheckInService) *CheckInHandler {
	return &CheckInHandler{
		service: service,
	}
}

// LookupBooking validates PNR and last name
// POST /api/webcheckin/lookup
func (h *CheckInHandler) LookupBooking(c *fiber.Ctx) error {
	type Request struct {
		PNR      string `json:"pnr"`
		LastName string `json:"lastName"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validation
	if req.PNR == "" || req.LastName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "PNR and last name are required"})
	}

	ctx := context.Background()
	booking, checkIn, err := h.service.LookupBooking(ctx, req.PNR, req.LastName)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": err.Error()})
	}

	// Determine check-in status
	var checkInStatus string
	if checkIn == nil {
		checkInStatus = "NOT_STARTED"
	} else {
		checkInStatus = string(checkIn.Status)
	}

	return c.JSON(fiber.Map{
		"booking": fiber.Map{
			"id":  booking.ID,
			"pnr": booking.PNR,
			"flight": fiber.Map{
				"id":            booking.Flight.ID,
				"code":          booking.Flight.Code,
				"source":        booking.Flight.Source,
				"destination":   booking.Flight.Destination,
				"departureTime": booking.Flight.DepartureTime,
			},
			"passenger": fiber.Map{
				"firstName":      booking.PassengerFirstName,
				"lastName":       booking.PassengerLastName,
				"passportNumber": booking.PassportNumber,
			},
			"initialSeatNo": booking.InitialSeatNo,
		},
		"checkinStatus": checkInStatus,
		"checkin":       checkIn,
	})
}

// GetSeats returns seat map for a flight
// GET /api/webcheckin/:pnr/seats?flightId=<uuid>
func (h *CheckInHandler) GetSeats(c *fiber.Ctx) error {
	// pnr is in route param but not used in this function
	flightIDStr := c.Query("flightId")

	if flightIDStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "flightId query parameter required"})
	}

	flightID, err := uuid.Parse(flightIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid flight ID"})
	}

	ctx := context.Background()
	seats, err := h.service.GetAvailableSeats(ctx, flightID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch seats"})
	}

	return c.JSON(fiber.Map{
		"flightId": flightID,
		"seats":    seats,
	})
}

// HoldSeat holds a seat for check-in
// POST /api/webcheckin/:pnr/hold-seat
func (h *CheckInHandler) HoldSeat(c *fiber.Ctx) error {
	pnr := c.Params("pnr")

	type Request struct {
		FlightID string `json:"flightId"`
		SeatNo   string `json:"seatNo"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	flightID, err := uuid.Parse(req.FlightID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid flight ID"})
	}

	ctx := context.Background()
	success, duration, err := h.service.HoldSeat(ctx, strings.ToUpper(pnr), flightID, req.SeatNo)
	if err != nil {
		statusCode := 409
		if strings.Contains(err.Error(), "not found") {
			statusCode = 404
		} else if strings.Contains(err.Error(), "occupied") {
			statusCode = 400
		}
		return c.Status(statusCode).JSON(fiber.Map{"error": err.Error()})
	}

	if !success {
		return c.Status(409).JSON(fiber.Map{"error": "Failed to hold seat"})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"holdDuration": duration,
	})
}

// CompleteCheckIn completes the check-in process
// POST /api/webcheckin/:pnr/complete
func (h *CheckInHandler) CompleteCheckIn(c *fiber.Ctx) error {
	pnr := c.Params("pnr")

	type Request struct {
		SeatNo        string  `json:"seatNo"`
		BaggageWeight float64 `json:"baggageWeight"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Lookup booking
	ctx := context.Background()
	booking, existingCheckIn, err := h.service.LookupBooking(ctx, pnr, "") // Empty lastName since we already validated
	if err != nil {
		// Use booking from session/context (simplified for demo)
		return c.Status(404).JSON(fiber.Map{"error": "Booking not found"})
	}

	// Check if already completed
	if existingCheckIn != nil && existingCheckIn.Status == models.CheckInStatusCompleted {
		return c.Status(400).JSON(fiber.Map{"error": "Already checked in"})
	}

	// If waiting for payment, resume check-in
	if existingCheckIn != nil && existingCheckIn.Status == models.CheckInStatusWaitingPayment {
		if !existingCheckIn.BaggageFeePaid {
			return c.Status(402).JSON(fiber.Map{
				"error":          "Payment required",
				"requiredAction": "PAYMENT",
				"details": fiber.Map{
					"currentWeight": existingCheckIn.BaggageWeight,
					"maxFree":       25.0,
					"excessWeight":  existingCheckIn.BaggageWeight - 25.0,
					"feeAmount":     existingCheckIn.BaggageFeeAmount,
				},
			})
		}

		// Resume check-in
		boardingPass, err := h.service.ResumeCheckIn(ctx, existingCheckIn.ID, booking, req.SeatNo)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"success":     true,
			"boardingPass": boardingPass,
		})
	}

	// New check-in
	checkIn, boardingPass, err := h.service.CompleteCheckIn(ctx, booking, req.SeatNo, req.BaggageWeight)
	if err != nil {
		// Check if payment required
		if strings.HasPrefix(err.Error(), "PAYMENT_REQUIRED:") {
			feeAmount := 0.0
			fmt.Sscanf(err.Error(), "PAYMENT_REQUIRED:%f", &feeAmount)

			return c.Status(402).JSON(fiber.Map{
				"error":          "Baggage overweight",
				"requiredAction": "PAYMENT",
				"checkInId":      checkIn.ID,
				"details": fiber.Map{
					"currentWeight": req.BaggageWeight,
					"maxFree":       25.0,
					"excessWeight":  req.BaggageWeight - 25.0,
					"feeAmount":     feeAmount,
				},
			})
		}

		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"checkin":      checkIn,
		"boardingPass": boardingPass,
	})
}

// ProcessBaggagePayment handles baggage fee payment
// POST /api/webcheckin/:pnr/baggage-payment
func (h *CheckInHandler) ProcessBaggagePayment(c *fiber.Ctx) error {
	// pnr is in route param but not used in this function

	type Request struct {
		CheckInID     string  `json:"checkInId"`
		FeeAmount     float64 `json:"feeAmount"`
		PaymentMethod string  `json:"paymentMethod"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	checkInID, err := uuid.Parse(req.CheckInID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid check-in ID"})
	}

	// Simulate payment processing
	if req.PaymentMethod != "SIMULATED" && req.PaymentMethod != "CARD" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payment method"})
	}

	ctx := context.Background()
	if err := h.service.ProcessBaggagePayment(ctx, checkInID, req.FeeAmount); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Payment processing failed"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Payment processed successfully. You may now complete check-in.",
	})
}

// SeedBookings seeds pre-existing bookings (development only)
// POST /api/webcheckin/admin/seed-bookings
func (h *CheckInHandler) SeedBookings(c *fiber.Ctx) error {
	type Request struct {
		Count int `json:"count"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Count == 0 {
		req.Count = 100
	}

	ctx := context.Background()
	if err := h.service.SeedBookings(ctx, req.Count); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Seeded %d bookings successfully", req.Count),
	})
}
