# API Specification - SkyHigh Digital Check-In System

## System Architecture

The SkyHigh Check-In System consists of two distinct backend services:
- **Booking Service (Backend 1):** Runs on port `8080`. Handles flight search, seat mapping for purchase, and initial flight bookings. Base URL: `/api/v1`
- **Web Check-In Service (Backend 2):** Runs on port `8081`. Handles PNR lookup, web check-in seat selection (120s holds), and baggage handling. Base URL: `/api/webcheckin`

---

## 1. Booking Service (Port 8080)

### 1.1 Get Available Flights
Retrieves a list of all flights available for initial booking.

- **URL**: `/flights`
- **Method**: `GET`
- **Response**: `200 OK`
```json
[
  {
    "id": "FL123",
    "destination": "New York",
    "departure_time": "2023-10-27T10:00:00Z",
    "plane_type": "Type A",
    "total_seats": 36
  }
]
```

### 2. Get Flight Details
Retrieves details for a specific flight.

- **URL**: `/flights/:flight_id`
- **Method**: `GET`
- **Response**: `200 OK`
```json
{
  "id": "FL123",
  "destination": "New York",
  "plane_type": "Type A",
  "rows": 6,
  "seats_per_row": 6
}
```

### 3. Get Seat Map
Retrieves the current seat map for a flight, including the status of each seat.

- **URL**: `/flights/:flight_id/seats`
- **Method**: `GET`
- **Response**: `200 OK`
```json
{
  "flight_id": "FL123",
  "seats": [
    {
      "seat_no": "1A",
      "status": "AVAILABLE", // AVAILABLE, HELD, CONFIRMED
      "row": 1,
      "column": "A"
    },
    {
      "seat_no": "1B",
      "status": "HELD",
      "held_expires_at": "2023-10-27T09:30:00Z" // Optional, if user has admin rights or for debugging
    }
  ]
}
```

### 4. Hold Seat
Attempts to place a temporary hold on a specific seat.

- **URL**: `/flights/:flight_id/seats/:seat_no/hold`
- **Method**: `POST`
- **Request Body**:
```json
{
  "user_id": "user-uuid-1234" // Identify the user holding the seat
}
```
- **Response**: `200 OK`
```json
{
  "message": "Seat held successfully",
  "expires_at": "2023-10-27T09:32:00Z", // 120 seconds from now
  "hold_token": "abc-123-token" // Token required to confirm booking
}
```
- **Error Responses**:
  - `409 Conflict`: Seat is already `HELD` or `CONFIRMED`.
  - `400 Bad Request`: Invalid seat number.

### 5. Confirm Check-In
Finalizes the check-in process, converting a `HELD` seat to `CONFIRMED`.

- **URL**: `/checkin/confirm`
- **Method**: `POST`
- **Request Body**:
```json
{
  "flight_id": "FL123",
  "seat_no": "1A",
  "hold_token": "abc-123-token",
  "passenger": {
    "first_name": "John",
    "last_name": "Doe",
    "passport": "A1234567"
  },
  "baggage": {
    "weight_kg": 20.5
  }
}
```
- **Response**: `200 OK`
```json
{
  "booking_reference": "REF123456",
  "status": "CONFIRMED",
  "seat_no": "1A"
}
```
- **Error Responses**:
  - `400 Bad Request`: Baggage weight exceeds 25kg (Payment Required logic would trigger here or separate endpoint).
  - `409 Conflict`: Hold expired or token invalid.
  - `422 Unprocessable Entity`: Data validation failed.

## Data Models

### Seat Authenticity
- `AVAILABLE`: Free to be held.
- `HELD`: Temporarily reserved (120s TTL).
- `CONFIRMED`: Permanently booked.

---

## 2. Web Check-In Service (Port 8081)

### 2.1 PNR Lookup
Verifies a booking by PNR and Last Name to begin the check-in process.

- **URL**: `/lookup`
- **Method**: `POST`
- **Request Body**:
```json
{
  "pnr": "ABC123",
  "last_name": "Doe"
}
```
- **Response**: `200 OK` (Returns booking and flight details)

### 2.2 Get Available Seats for Check-In
Retrieves available seats specifically for the check-in process.

- **URL**: `/:pnr/seats`
- **Method**: `GET`
- **Response**: `200 OK` (Array of available seats, excluding already checked-in seats)

### 2.3 Hold Seat for Check-In
Places a hardware locking 120-second hold on a specific seat for web check-in.

- **URL**: `/:pnr/hold-seat`
- **Method**: `POST`
- **Request Body**:
```json
{
  "seat_no": "6B"
}
```
- **Response**: `200 OK` (Returns expiration time and hold token)

### 2.4 Complete Check-in
Finalizes check-in, verifies baggage weight, and returns the boarding pass.

- **URL**: `/:pnr/complete`
- **Method**: `POST`
- **Request Body**:
```json
{
  "baggage_weight": 20
}
```
- **Response**: `200 OK` (Check-in complete, generates boarding pass properties)
- **Response (If Overweight >25kg)**: `402 Payment Required`

### 2.5 Process Baggage Payment
Records payment for excess baggage fees.

- **URL**: `/:pnr/baggage-payment`
- **Method**: `POST`
- **Request Body**:
```json
{
  "amount": 50.00
}
```
- **Response**: `200 OK` (Payment successful, fee_paid set to true)
