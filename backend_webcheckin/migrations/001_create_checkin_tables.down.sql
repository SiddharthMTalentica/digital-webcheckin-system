-- Down Migration: Rollback checkin tables

-- Drop indexes
DROP INDEX IF EXISTS idx_seats_checkin;
DROP INDEX IF EXISTS idx_checkins_status;
DROP INDEX IF EXISTS idx_checkins_booking;
DROP INDEX IF EXISTS idx_bookings_flight_passenger;
DROP INDEX IF EXISTS idx_bookings_pnr;

-- Remove checkin_id column from seats
ALTER TABLE seats DROP COLUMN IF EXISTS checkin_id;

-- Drop tables (CASCADE to remove foreign key constraints)
DROP TABLE IF EXISTS checkins CASCADE;
DROP TABLE IF EXISTS bookings CASCADE;

-- Drop enum type
DROP TYPE IF EXISTS checkin_status;
