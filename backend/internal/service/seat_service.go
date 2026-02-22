package service

import (
	"context"
	"digital-checkin/internal/core"
	"digital-checkin/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SeatService struct {
	Repo  *repository.Repository
	Redis *redis.Client
}

func NewSeatService(repo *repository.Repository, rdb *redis.Client) *SeatService {
	return &SeatService{
		Repo:  repo,
		Redis: rdb,
	}
}

// GetFlightSeats merges DB state (CONFIRMED) with Redis state (HELD).
func (s *SeatService) GetFlightSeats(ctx context.Context, flightID uuid.UUID) ([]core.Seat, error) {
	// 1. Get Seats from DB
	seats, err := s.Repo.GetSeatsByFlightID(ctx, flightID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch seats from DB: %w", err)
	}

	// 2. Check Redis for any active HOLDS
	pipe := s.Redis.Pipeline()
    cmds := make(map[string]*redis.IntCmd)

	for _, seat := range seats {
		if !seat.IsBooked {
			key := fmt.Sprintf("hold:%s:%s", flightID, seat.SeatNo)
			cmds[seat.SeatNo] = pipe.Exists(ctx, key)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check redis holds: %w", err)
	}

	// 3. Merge Status
	for i, seat := range seats {
		if seat.IsBooked {
			seats[i].Status = core.SeatConfirmed
			continue
		}

		if cmd, ok := cmds[seat.SeatNo]; ok {
            if cmd.Val() > 0 {
                seats[i].Status = core.SeatHeld
            } else {
                seats[i].Status = core.SeatAvailable
            }
        }
	}

	return seats, nil
}

func (s *SeatService) GetAllFlights(ctx context.Context) ([]core.Flight, error) {
	return s.Repo.GetAllFlights(ctx)
}

// HoldSeat attempts to hold a seat for 120 seconds.
func (s *SeatService) HoldSeat(ctx context.Context, flightID uuid.UUID, seatNo string, userID string) (string, time.Time, error) {
    isBooked, err := s.Repo.IsSeatBooked(ctx, flightID, seatNo)
    if err != nil {
        return "", time.Time{}, fmt.Errorf("failed to check seat status: %w", err)
    }
    if isBooked {
        return "", time.Time{}, fmt.Errorf("seat %s is already booked", seatNo)
    }

    key := fmt.Sprintf("hold:%s:%s", flightID, seatNo)
    expiration := 45 * time.Second
    
    success, err := s.Redis.SetNX(ctx, key, userID, expiration).Result()
    if err != nil {
        return "", time.Time{}, fmt.Errorf("redis error: %w", err)
    }
    
    if !success {
        return "", time.Time{}, fmt.Errorf("seat %s is currently held by another user", seatNo)
    }
    
    expiresAt := time.Now().Add(expiration)
    return key, expiresAt, nil
}

type CheckInRequest struct {
    FlightID uuid.UUID
    SeatNo   string
    UserID   string
     PassengerFirstName string
    PassengerLastName  string
    PassengerPassport  string
    BaggageWeight       float64
}

// ConfirmCheckIn finalizes the booking
func (s *SeatService) ConfirmCheckIn(ctx context.Context, req CheckInRequest) (*core.Booking, error) {
    // 1. Verify Hold Logic ONLY if seat is selected
    if req.SeatNo != "" {
        // In a real system, we'd check if the UserID matches the one who holds the lock.
        key := fmt.Sprintf("hold:%s:%s", req.FlightID, req.SeatNo)
        holderID, err := s.Redis.Get(ctx, key).Result()
        
        if err == redis.Nil {
            return nil, fmt.Errorf("seat hold expired or invalid")
        } else if err != nil {
             return nil, fmt.Errorf("redis error: %w", err)
        }
        
        if holderID != req.UserID {
            return nil, fmt.Errorf("seat is held by another user")
        }
    }
    
    // 2. Validate Baggage
    if req.BaggageWeight > 25.0 {
        return nil, fmt.Errorf("baggage weight exceeds 25kg limit. Payment required")
    }
    
    // 3. Create Booking in DB (Transactional)
    // This should ideally be a TX: Insert Booking + Update Seat IsBooked = TRUE
    booking, err := s.Repo.CreateBooking(ctx, req.FlightID, req.SeatNo, req.PassengerFirstName, req.PassengerLastName, req.PassengerPassport)
    if err != nil {
        return nil, fmt.Errorf("failed to create booking: %w", err)
    }
    
    // 4. Release Hold (Optional, but good practice to clear Redis)
    if req.SeatNo != "" {
        key := fmt.Sprintf("hold:%s:%s", req.FlightID, req.SeatNo)
        s.Redis.Del(ctx, key)
    }
    
    return booking, nil
}
