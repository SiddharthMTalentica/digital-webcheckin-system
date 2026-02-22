# Workflow Design & Database Schema

## 1. Flow Diagrams

### Overall Web Check-In Flow

```text
+-----------------------------------+
| 1. Passenger Enters PNR & Name    |
+-----------------------------------+
                 |
                 v
           [Valid PNR?] ----(NO)---> (Show Error)
                 |
               (YES)
                 |
                 v
        [Already Checked In?] --(YES)-> [Show Boarding Pass]
                 |
               (NO)
                 |
                 v
        [Show Flight Details]
                 |
                 v
        [Passenger Selects Seat]
                 |
                 v
          [Seat Available?] --(NO)--> (Show Error & Refresh)
                 |
               (YES)
                 |
                 v
+-----------------------------------+
| 2. Acquire 120s Lock in Redis     |
+-----------------------------------+
                 |
                 v
      [Enter Baggage Weight]
                 |
                 v
         [Weight > 25kg?] --(YES)---+
                 |                  |
               (NO)           [Pause Check-in]
                 |            [Require $10/kg]
                 |                  |
                 |             (User Pays)
                 |<-----------------+
                 |
                 v
         [Hold Expired?] --(YES)---> (Release Seat, Show Error)
                 |
               (NO)
                 |
                 v
+-----------------------------------+
| 3. Save Check-in to PostgreSQL DB |
+-----------------------------------+
                 |
                 v
+-----------------------------------+
| 4. Generate & Show Boarding Pass  |
+-----------------------------------+
```

### Seat Hold Process

```text
  User               API               Redis                DB
   |                  |                  |                  |
   |---- Select Seat >|                  |                  |
   |   (POST /hold)   |--- Is Booked? --------------------->|
   |                  |<------ No --------------------------|
   |                  |                  |                  |
   |                  |--- SET NX key -->|                  |
   |                  |  (TTL 120s)      |                  |
   |                  |                  |                  |
   |                  |  [Lock Acquired] |                  |
   |                  |<-- Success (1) --|                  |
   |<-- Seat Held ----|                  |                  |
   |  (Token Returned)|                  |                  |
   |                  |                  |                  |
   |                  |   [Lock Failed]  |                  |
   |                  |<---- Fail (0) ---|                  |
   |<-- Error: Taken -|                  |                  |
   |                  |                  |                  |
```

### Check-in Confirmation Process

```text
  User               API               Redis                DB
   |                  |                  |                  |
   |---- Confirm ---->|                  |                  |
   |    Check-in      |-- Check Token -->|                  |
   |                  |                  |                  |
   |                  |   [Valid Lock]   |                  |
   |                  |                  |                  |
   |                  |(Validate Baggage)|                  |
   |                  |                  |                  |
   |                  |  [Baggage OK]    |                  |
   |                  |-- Start TX ------------------------>|
   |                  |                  |                  |
   |                  |                  |  (Insert Booking)|
   |                  |                  |                  |
   |                  |                  |   (Update Seat)  |
   |                  |<--------------------- Commit TX ----|
   |                  |--- DEL Lock ---->|                  |
   |<-- Successful ---|                  |                  |
   |                  |                  |                  |
   |                  |  [Overweight]    |                  |
   |<-- Payment Req --|                  |                  |
   |                  |                  |                  |
   |                  | [Invalid Lock]   |                  |
   |<-- Expired Error-|                  |                  |
   |                  |                  |                  |
```

## 2. Database Schema

```text
 +--------------------+        +--------------------+
 | FLIGHTS            | 1    * | SEATS              |
 |--------------------|--------|--------------------|
 | id (PK)            |        | id (PK)            |
 | code               |        | flight_id (FK)     |
 | destination        |        | seat_no            |
 | departure_time     |        | row_num            |
 | plane_type         |        | col_num            |
 +--------------------+        | category           |
           | 1                 | is_booked          |
           |                   +--------------------+
           |                             | 1
           |                             |
           | *                           | 0..1
 +--------------------+                  |
 | BOOKINGS           |                  |
 |--------------------|                  |
 | id (PK)            |                  |
 | flight_id (FK)     |<-----------------+
 | seat_id (FK)       |
 | passenger_name     |
 | booking_ref        |
 | status             |
 +--------------------+
```

### Table Descriptions

- **FLIGHTS**: Stores flight metadata.
- **SEATS**: Stores valid seats for a flight. `is_booked` is the source of truth for confirmed bookings.
- **BOOKINGS**: Records passenger details and links to a specific seat.
