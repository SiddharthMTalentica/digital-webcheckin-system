-- Migration: Create bookings and checkins tables for web check-in system
-- Up Migration

-- Create bookings table (pre-existing flight bookings)
CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pnr VARCHAR(6) UNIQUE NOT NULL,
    flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    passenger_first_name VARCHAR(100) NOT NULL,
    passenger_last_name VARCHAR(100) NOT NULL,
    passport_number VARCHAR(50),
    email VARCHAR(255),
    phone VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for bookings
CREATE INDEX idx_bookings_pnr ON bookings(pnr);
CREATE INDEX idx_bookings_flight_passenger ON bookings(flight_id, passenger_last_name);

-- Create check-in status enum
DO $$ BEGIN
    CREATE TYPE checkin_status AS ENUM ('IN_PROGRESS', 'WAITING_PAYMENT', 'COMPLETED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create checkins table
CREATE TABLE IF NOT EXISTS checkins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL UNIQUE REFERENCES bookings(id) ON DELETE CASCADE,
    seat_no VARCHAR(10) NOT NULL,
    baggage_weight DECIMAL(5, 2) DEFAULT 0,
    baggage_fee_paid BOOLEAN DEFAULT FALSE,
    baggage_fee_amount DECIMAL(10, 2) DEFAULT 0,
    status checkin_status NOT NULL DEFAULT 'IN_PROGRESS',
    checked_in_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

-- Create indexes for checkins
CREATE INDEX idx_checkins_booking ON checkins(booking_id);
CREATE INDEX idx_checkins_status ON checkins(status);

-- Add checkin_id column to seats table (if not exists)
DO $$ BEGIN
    ALTER TABLE seats ADD COLUMN checkin_id UUID REFERENCES checkins(id);
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

-- Create index on checkin_id
CREATE INDEX IF NOT EXISTS idx_seats_checkin ON seats(checkin_id);
