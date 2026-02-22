# Project Structure

This document outlines the folder structure and key modules of the SkyHigh Core backend.

```text
digital_checking_system/
├── backend/                  # Source code for the Booking Go application (Port 8080)
│   ├── cmd/
│   │   └── server/           # Application entry point (main.go)
│   ├── internal/
│   │   ├── api/              # HTTP Handlers (Controllers) & Routing logic
│   │   ├── core/             # Domain Models (Structs) & Seeding logic
│   │   ├── repository/       # Data Access Layer (PostgreSQL implementation)
│   │   └── service/          # Business Logic Layer (Seat Management, Redis Locking)
│   ├── pkg/
│   │   ├── config/           # Configuration management (Env vars)
│   │   ├── db/               # Database connection helper
│   │   └── redis/            # Redis connection helper
│   ├── migrations/           # SQL Migration files
│   ├── go.mod                # Go module definition
│   └── go.sum                # Go dependencies checksum
├── backend_webcheckin/       # Source code for the Web Check-In application (Port 8081)
│   ├── cmd/
│   │   └── main.go           # Application entry point (main.go)
│   ├── internal/
│   │   ├── handler/          # HTTP Handlers & Fiber configuration
│   │   ├── repository/       # Read/Write DB Operations
│   │   └── service/          # Web Check-in Business Logic & Baggage Check
│   ├── migrations/           # SQL specific to Checkins
│   ├── go.mod                # Go module definition
│   └── go.sum                # Go dependencies checksum
├── frontend/                 # React UI serving both systems
├── docker-compose.yml        # Infrastructure setup (Postgres, Redis, Backends)
├── PRD.md                    # Product Requirements Document
├── README.md                 # Project Overview & Setup
├── API-SPECIFICATION.md      # Dual System API Documentation
└── tasks.md                  # Development Task Tracking
```

## Key Modules

### `internal/api`
Contains the HTTP handlers that process incoming requests. It parses JSON payloads, validates inputs, calls the Service layer, and formats responses.

### `internal/service`
Contains the core business logic.
- **SeatService**: Orchestrates the check-in flow. It communicates with the Repository for persistent data and Redis for ephemeral state (locks/holds). It implements the "Hard Guarantee" logic for seat reservation.

### `internal/repository`
Handles direct interaction with the PostgreSQL database. It performs CRUD operations for Flights, Seats, and Bookings. It also manages database transactions for the final booking confirmation.

### `pkg/redis` & `pkg/db`
Reusable packages for initializing connections to infrastructure components. separating these ensures clean dependency injection and easier testing.
