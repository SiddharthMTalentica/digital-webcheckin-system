CREATE TABLE IF NOT EXISTS flights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    source VARCHAR(100) NOT NULL,
    destination VARCHAR(100) NOT NULL,
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    plane_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id UUID NOT NULL REFERENCES flights(id) ON DELETE CASCADE,
    seat_no VARCHAR(10) NOT NULL,
    row_num INT NOT NULL,
    col_num VARCHAR(5) NOT NULL,
    category VARCHAR(20) DEFAULT 'STANDARD', -- Helper for different seat types if needed
    is_booked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(flight_id, seat_no)
);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flight_id UUID NOT NULL REFERENCES flights(id),
    seat_id UUID NOT NULL REFERENCES seats(id),
    passenger_first_name VARCHAR(100) NOT NULL,
    passenger_last_name VARCHAR(100) NOT NULL,
    passenger_passport VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'CONFIRMED',
    booking_reference VARCHAR(20) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_seats_flight_id ON seats(flight_id);
CREATE INDEX idx_bookings_flight_id ON bookings(flight_id);
CREATE INDEX idx_bookings_seat_id ON bookings(seat_id);
