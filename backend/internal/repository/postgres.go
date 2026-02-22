package repository

import (
	"context"
	"database/sql"
	"digital-checkin/internal/core"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetFlightByID(ctx context.Context, flightID uuid.UUID) (*core.Flight, error) {
	query := `
		SELECT id, code, source, destination, departure_time, plane_type, created_at, updated_at
		FROM flights
		WHERE id = $1
	`
	var f core.Flight
	err := r.DB.QueryRowContext(ctx, query, flightID).Scan(
		&f.ID, &f.Code, &f.Source, &f.Destination, &f.DepartureTime, &f.PlaneType, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &f, nil
}

func (r *Repository) GetSeatsByFlightID(ctx context.Context, flightID uuid.UUID) ([]core.Seat, error) {
	query := `
		SELECT id, flight_id, seat_no, row_num, col_num, category, is_booked, created_at, updated_at
		FROM seats
		WHERE flight_id = $1
		ORDER BY row_num, col_num
	`
	rows, err := r.DB.QueryContext(ctx, query, flightID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []core.Seat
	for rows.Next() {
		var s core.Seat
		err := rows.Scan(
			&s.ID, &s.FlightID, &s.SeatNo, &s.RowNum, &s.ColNum, &s.Category, &s.IsBooked, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if s.IsBooked {
			s.Status = core.SeatConfirmed
		} else {
			s.Status = core.SeatAvailable
		}
		seats = append(seats, s)
	}
	return seats, nil
}

func (r *Repository) GetAllFlights(ctx context.Context) ([]core.Flight, error) {
	query := `SELECT id, code, source, destination, departure_time, plane_type, created_at, updated_at FROM flights`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []core.Flight
	for rows.Next() {
		var f core.Flight
		if err := rows.Scan(&f.ID, &f.Code, &f.Source, &f.Destination, &f.DepartureTime, &f.PlaneType, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		flights = append(flights, f)
	}
	return flights, nil
}

func (r *Repository) IsSeatBooked(ctx context.Context, flightID uuid.UUID, seatNo string) (bool, error) {
    var isBooked bool
    query := `SELECT is_booked FROM seats WHERE flight_id = $1 AND seat_no = $2`
    err := r.DB.QueryRowContext(ctx, query, flightID, seatNo).Scan(&isBooked)
    if err != nil {
        if err == sql.ErrNoRows {
            return false, fmt.Errorf("seat not found")
        }
        return false, err
    }
    return isBooked, nil
}

// CreateBooking performs a transaction:
// 1. Get Seat ID
// 2. Insert Booking
// 3. Update Seat -> is_booked = true
func (r *Repository) CreateBooking(ctx context.Context, flightID uuid.UUID, seatNo, fName, lName, passport string) (*core.Booking, error) {
    tx, err := r.DB.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // 1. Get Seat ID and lock row if seat is selected
    var seatID *uuid.UUID

    if seatNo != "" {
        var id uuid.UUID
        var isBooked bool
        err = tx.QueryRowContext(ctx, `SELECT id, is_booked FROM seats WHERE flight_id = $1 AND seat_no = $2 FOR UPDATE`, flightID, seatNo).Scan(&id, &isBooked)
        if err != nil {
            return nil, err
        }
        
        if isBooked {
            return nil, fmt.Errorf("seat already booked in DB (check failed)")
        }
        seatID = &id
    }

    // 2. Insert Booking
    bookingID := uuid.New()
    // Use last 8 digits of UnixNano for brevity + random component or just random string
    // REF-12345678 (12 chars)
    bookingRef := fmt.Sprintf("REF-%d", time.Now().UnixNano()%100000000)
    
	// Generate PNR
	pnr := generatePNR()

    _, err = tx.ExecContext(ctx, `
        INSERT INTO bookings (id, flight_id, seat_id, passenger_first_name, passenger_last_name, passenger_passport, status, booking_reference, pnr)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `, bookingID, flightID, seatID, fName, lName, passport, "CONFIRMED", bookingRef, pnr)
    if err != nil {
        return nil, err
    }

    // 3. Update Seat if selected
    if seatID != nil {
        _, err = tx.ExecContext(ctx, `UPDATE seats SET is_booked = TRUE, updated_at = NOW() WHERE id = $1`, seatID)
        if err != nil {
            return nil, err
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return &core.Booking{
        ID: bookingID,
        FlightID: flightID,
        SeatID: seatID,
        PassengerFirstName: fName,
        PassengerLastName: lName, 
        PassengerPassport: passport,
        Status: "CONFIRMED",
        BookingReference: bookingRef,
		PNR: pnr,
        CreatedAt: time.Now(),
    }, nil
}

func generatePNR() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Exclude I, O, 0, 1 for clarity
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
