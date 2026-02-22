ALTER TABLE bookings ADD COLUMN pnr VARCHAR(6);
CREATE UNIQUE INDEX idx_bookings_pnr ON bookings(pnr);
