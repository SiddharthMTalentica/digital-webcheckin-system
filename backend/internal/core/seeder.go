package core

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type Seeder struct {
	DB *sql.DB
}

func NewSeeder(db *sql.DB) *Seeder {
	return &Seeder{DB: db}
}

func (s *Seeder) Seed() error {
	ctx := context.Background()

	// Check if flights exist
	var count int
	err := s.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM flights").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Data already seeded, skipping...")
		return nil
	}

	log.Println("Seeding data for next 2 months...")

	// Generate flights for the next 60 days (2 months)
	baseDate := time.Now()
	
	for day := 0; day < 60; day++ {
		departureDate := baseDate.Add(time.Duration(day) * 24 * time.Hour)
		
		// Create unique flight codes for each day
		dayStr := fmt.Sprintf("%02d", day)
		
		// 1. Flight Type A (Small - 6 seats/row)
		// 6 rows * 6 cols = 36 seats
		flightCode1 := fmt.Sprintf("FL001-D%s", dayStr)
		err = s.createFlightWithSeatsAndDate(ctx, flightCode1, "Pune", "New York", "Type A", 6, 6, departureDate)
		if err != nil {
			return err
		}

		// 2. Flight Type B (4 seats/row * 18 rows)
		flightCode2 := fmt.Sprintf("FL002-D%s", dayStr)
		err = s.createFlightWithSeatsAndDate(ctx, flightCode2, "Los Angeles", "London", "Type B", 18, 4, departureDate)
		if err != nil {
			return err
		}

		// 3. Flight Type C (International - 12 seats/row)
		// 10 rows * 12 cols = 120 seats
		flightCode3 := fmt.Sprintf("FL003-D%s", dayStr)
		err = s.createFlightWithSeatsAndDate(ctx, flightCode3, "Mumbai", "Tokyo", "Type C", 10, 12, departureDate)
		if err != nil {
			return err
		}
	}

	log.Printf("Seeding completed successfully. Created %d flights for next 60 days.", 60*3)
	return nil
}

func (s *Seeder) createFlightWithSeatsAndDate(ctx context.Context, code, source, dest, pType string, rows, cols int, depTime time.Time) error {
	flightID := uuid.New()

	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO flights (id, code, source, destination, departure_time, plane_type)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, flightID, code, source, dest, depTime, pType)

	if err != nil {
		return fmt.Errorf("failed to insert flight %s: %v", code, err)
	}

	// Generate Seats
	stmt, err := s.DB.PrepareContext(ctx, `
		INSERT INTO seats (flight_id, seat_no, row_num, col_num, category)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	colLetters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for r := 1; r <= rows; r++ {
		for c := 0; c < cols; c++ {
			colChar := string(colLetters[c])
			seatNo := fmt.Sprintf("%d%s", r, colChar)
			category := "STANDARD"
			
			// Simple logic: First 2 rows are PREMIUM
			if r <= 2 {
				category = "PREMIUM"
			}

			_, err := stmt.ExecContext(ctx, flightID, seatNo, r, colChar, category)
			if err != nil {
				return fmt.Errorf("failed to insert seat %s: %v", seatNo, err)
			}
		}
	}
	
	log.Printf("Created Flight %s (%s) from %s to %s with %d seats.", code, pType, source, dest, rows*cols)
	return nil
}
