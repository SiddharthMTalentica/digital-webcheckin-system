package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend_webcheckin/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewRepository(dbHost, dbUser, dbPassword, dbName string, dbPort int, redisHost string, redisPort int, redisPassword string) (*Repository, error) {
	// PostgreSQL Connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Redis Connection
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Repository{
		DB:    db,
		Redis: rdb,
	}, nil
}

// FindBookingByPNR validates PNR and optionally last name, returns booking with flight details.
// Searches webcheckin_bookings first, then falls back to the bookings table (from the booking backend)
// and syncs the record into webcheckin_bookings if found.
func (r *Repository) FindBookingByPNR(ctx context.Context, pnr, lastName string) (*models.Booking, error) {
	var booking models.Booking
	
	query := r.DB.WithContext(ctx).Preload("Flight")
	
	if lastName != "" {
		query = query.Where("UPPER(pnr) = ? AND LOWER(passenger_last_name) = ?", 
			strings.ToUpper(pnr), strings.ToLower(lastName))
	} else {
		query = query.Where("UPPER(pnr) = ?", strings.ToUpper(pnr))
	}
	
	result := query.First(&booking)
	
	if result.Error == nil {
		return &booking, nil
	}
	
	if result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}
	
	// Not found in webcheckin_bookings — try the main bookings table
	var mainPNR, mainFlightID, mainFirstName, mainLastName, mainPassport string
	
	fallbackQuery := `SELECT pnr, flight_id, passenger_first_name, passenger_last_name, COALESCE(passenger_passport, '') FROM bookings WHERE UPPER(pnr) = ?`
	args := []interface{}{strings.ToUpper(pnr)}
	
	if lastName != "" {
		fallbackQuery += ` AND LOWER(passenger_last_name) = ?`
		args = append(args, strings.ToLower(lastName))
	}
	fallbackQuery += ` LIMIT 1`
	
	row := r.DB.WithContext(ctx).Raw(fallbackQuery, args...).Row()
	if err := row.Scan(&mainPNR, &mainFlightID, &mainFirstName, &mainLastName, &mainPassport); err != nil {
		return nil, fmt.Errorf("booking not found or name mismatch")
	}
	
	// Sync into webcheckin_bookings so checkins FK works
	err := r.DB.WithContext(ctx).Exec(
		`INSERT INTO webcheckin_bookings (pnr, flight_id, passenger_first_name, passenger_last_name, passport_number)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (pnr) DO NOTHING`,
		mainPNR, mainFlightID, mainFirstName, mainLastName, mainPassport,
	).Error
	if err != nil {
		return nil, fmt.Errorf("failed to sync booking: %w", err)
	}
	
	// Now re-query from webcheckin_bookings to get proper GORM model with Flight preloaded
	reQuery := r.DB.WithContext(ctx).Preload("Flight").Where("UPPER(pnr) = ?", strings.ToUpper(pnr))
	result = reQuery.First(&booking)
	if result.Error != nil {
		return nil, fmt.Errorf("booking not found or name mismatch")
	}
	
	return &booking, nil
}

// FindInitialSeatByPNR finds the initially selected seat for a booking
func (r *Repository) FindInitialSeatByPNR(ctx context.Context, pnr string) (string, error) {
	var seatNo string
	query := `
		SELECT s.seat_no 
		FROM bookings b 
		JOIN seats s ON b.seat_id = s.id 
		WHERE b.pnr = $1 AND b.seat_id IS NOT NULL
	`
	err := r.DB.WithContext(ctx).Raw(query, pnr).Scan(&seatNo).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return "", err
	}
	return seatNo, nil
}

// FindCheckInByBookingID retrieves existing check-in for a booking
func (r *Repository) FindCheckInByBookingID(ctx context.Context, bookingID uuid.UUID) (*models.CheckIn, error) {
	var checkIn models.CheckIn
	result := r.DB.WithContext(ctx).Where("booking_id = ?", bookingID).First(&checkIn)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // No check-in yet, not an error
		}
		return nil, result.Error
	}
	
	return &checkIn, nil
}

// GetAvailableSeats returns seats for a flight, excluding checked-in seats
func (r *Repository) GetAvailableSeats(ctx context.Context, flightID uuid.UUID) ([]models.Seat, error) {
	var seats []models.Seat
	result := r.DB.WithContext(ctx).
		Where("flight_id = ?", flightID).
		Order("seat_no ASC").
		Find(&seats)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return seats, nil
}

// GetSeatByFlightAndSeatNo retrieves a specific seat
func (r *Repository) GetSeatByFlightAndSeatNo(ctx context.Context, flightID uuid.UUID, seatNo string) (*models.Seat, error) {
	var seat models.Seat
	result := r.DB.WithContext(ctx).
		Where("flight_id = ? AND seat_no = ?", flightID, seatNo).
		First(&seat)
	
	if result.Error != nil {
		return nil, result.Error
	}
	
	return &seat, nil
}

// HoldSeat holds a seat using Redis with TTL
func (r *Repository) HoldSeat(ctx context.Context, pnr, seatNo string, flightID uuid.UUID, duration int) (bool, error) {
	key := fmt.Sprintf("checkin:hold:%s:%s", pnr, seatNo)
	expiration := time.Duration(duration) * time.Second
	
	// Use SetNX to ensure atomic hold (only succeeds if key doesn't exist)
	success, err := r.Redis.SetNX(ctx, key, flightID.String(), expiration).Result()
	if err != nil {
		return false, err
	}
	
	return success, nil
}

// ReleaseSeatHold releases a Redis hold
func (r *Repository) ReleaseSeatHold(ctx context.Context, pnr, seatNo string) error {
	key := fmt.Sprintf("checkin:hold:%s:%s", pnr, seatNo)
	return r.Redis.Del(ctx, key).Err()
}

// CreateCheckIn creates a new check-in record
func (r *Repository) CreateCheckIn(ctx context.Context, checkIn *models.CheckIn) error {
	return r.DB.WithContext(ctx).Create(checkIn).Error
}

// UpdateCheckInStatus updates the status of a check-in
func (r *Repository) UpdateCheckInStatus(ctx context.Context, id uuid.UUID, status models.CheckInStatus) error {
	updates := map[string]interface{}{
		"status": status,
	}
	
	if status == models.CheckInStatusCompleted {
		now := time.Now()
		updates["completed_at"] = &now
	}
	
	return r.DB.WithContext(ctx).
		Model(&models.CheckIn{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateCheckInBaggageFee records baggage fee payment
func (r *Repository) UpdateCheckInBaggageFee(ctx context.Context, id uuid.UUID, feeAmount float64) error {
	return r.DB.WithContext(ctx).
		Model(&models.CheckIn{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"baggage_fee_paid":   true,
			"baggage_fee_amount": feeAmount,
		}).Error
}

// UpdateSeatWithCheckIn links a seat to a check-in
func (r *Repository) UpdateSeatWithCheckIn(ctx context.Context, seatID, checkInID uuid.UUID) error {
	return r.DB.WithContext(ctx).
		Model(&models.Seat{}).
		Where("id = ?", seatID).
		Updates(map[string]interface{}{
			"checkin_id": checkInID,
			"is_booked":  true,
		}).Error
}

// ReleaseOldSeat releases an old seat that was previously booked
func (r *Repository) ReleaseOldSeat(ctx context.Context, seatID uuid.UUID) error {
	return r.DB.WithContext(ctx).
		Model(&models.Seat{}).
		Where("id = ?", seatID).
		Updates(map[string]interface{}{
			"checkin_id": nil,
			"is_booked":  false,
		}).Error
}

// UpdateBookingSeat updates the seat_id in the bookings table
func (r *Repository) UpdateBookingSeat(ctx context.Context, pnr string, newSeatID uuid.UUID) error {
	return r.DB.WithContext(ctx).Exec(
		`UPDATE bookings SET seat_id = $1 WHERE UPPER(pnr) = $2`, 
		newSeatID, strings.ToUpper(pnr),
	).Error
}

// Close closes database and Redis connections
func (r *Repository) Close() {
	if r.Redis != nil {
		r.Redis.Close()
	}
	if r.DB != nil {
		sqlDB, _ := r.DB.DB()
		sqlDB.Close()
	}
}
