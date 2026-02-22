# System Architecture - Digital Check-In System

> **Version**: 2.0 (Dual System Architecture)  
> **Last Updated**: 2026-02-16

---

## System Overview

The Digital Check-In System now consists of **two independent subsystems** running on separate routes, sharing common infrastructure.

---

## High-Level Architecture

```mermaid
graph TB
    subgraph "User Interface (Frontend - React)"
        direction TB
        Browser[Web Browser]
        
        subgraph "Route: /"
            Home[Homepage Component<br/>System Selector]
        end
        
        subgraph "Route: /flight-booking-system"
            BookingUI[Flight Booking UI<br/>Date Picker → Flight List → Seat Map]
        end
        
        subgraph "Route: /web-check-in"
            CheckInUI[Web Check-In UI<br/>PNR Lookup → Seat Selection → Boarding Pass]
        end
    end
    
    subgraph "Backend Services (Go + Fiber)"
        direction LR
        
        subgraph "Booking Service :8080"
            BookingAPI[Flight Booking API<br/>- Search Flights<br/>- Book Tickets<br/>- Seat Management]
        end
        
        subgraph "Check-In Service :8081"
            CheckInAPI[Web Check-In API<br/>- PNR Lookup<br/>- Seat Hold (120s)<br/>- Complete Check-In]
        end
    end
    
    subgraph "Data Layer"
        direction TB
        
        DB[(PostgreSQL<br/>- Flights<br/>- Seats<br/>- Bookings<br/>- CheckIns)]
        
        Cache[(Redis<br/>- Seat Holds<br/>- TTL: 45s (booking)<br/>- TTL: 120s (check-in))]
    end
    
    Browser --> Home
    Home -->|Book Flight| BookingUI
    Home -->|Check-In| CheckInUI
    
    BookingUI --> BookingAPI
    CheckInUI --> CheckInAPI
    
    BookingAPI --> DB
    BookingAPI --> Cache
    CheckInAPI --> DB
    CheckInAPI --> Cache
    
    style Home fill:#e3f2fd
    style BookingUI fill:#fff3e0
    style CheckInUI fill:#e8f5e9
    style BookingAPI fill:#fff3e0
    style CheckInAPI fill:#e8f5e9
```

---

## Component Details

### Frontend Architecture

```mermaid
graph LR
    subgraph "Frontend Components"
        Main[main.jsx<br/>React Router Setup]
        
        Main --> Homepage[Homepage.jsx<br/>System Selector]
        Main --> BookingApp[FlightBookingApp.jsx<br/>Existing System]
        Main --> CheckInApp[WebCheckInApp.jsx<br/>New System]
        
        subgraph "Booking Components"
            BookingApp --> FlightList[FlightList]
            BookingApp --> SeatMapBooking[SeatMap]
            BookingApp --> PassengerForm[PassengerForm]
        end
        
        subgraph "Check-In Components"
            CheckInApp --> PNRLookup[PNRLookup]
            CheckInApp --> FlightConfirmation[FlightConfirmation]
            CheckInApp --> SeatMapCheckIn[SeatMap<br/>Reused]
            CheckInApp --> BaggageForm[BaggageForm]
            CheckInApp --> PaymentModal[PaymentModal]
            CheckInApp --> BoardingPass[BoardingPass]
        end
        
        FlightList --> API1[Booking API]
        SeatMapBooking --> API1
        PassengerForm --> API1
        
        PNRLookup --> API2[Check-In API]
        FlightConfirmation --> API2
        SeatMapCheckIn --> API2
        BaggageForm --> API2
        PaymentModal --> API2
        BoardingPass --> API2
    end
    
    style BookingApp fill:#fff3e0
    style CheckInApp fill:#e8f5e9
```

### Backend Microservices

```mermaid
graph TB
    subgraph "Booking Service (Port 8080)"
        direction TB
        BookingHandler[HTTP Handlers<br/>Fiber Router]
        BookingService[Business Logic<br/>Flight Search, Bookings]
        BookingRepo[Repository Layer<br/>PostgreSQL Access]
        BookingCache[Redis Client<br/>45s Hold]
        
        BookingHandler --> BookingService
        BookingService --> BookingRepo
        BookingService --> BookingCache
    end
    
    subgraph "Check-In Service (Port 8081)"
        direction TB
        CheckInHandler[HTTP Handlers<br/>Fiber Router]
        CheckInService[Business Logic<br/>PNR Validation, Check-In]
        CheckInRepo[Repository Layer<br/>PostgreSQL Access]
        CheckInCache[Redis Client<br/>120s Hold]
        
        CheckInHandler --> CheckInService
        CheckInService --> CheckInRepo
        CheckInService --> CheckInCache
    end
    
    BookingRepo -.->|Shared DB| DB[(PostgreSQL)]
    CheckInRepo -.->|Shared DB| DB
    
    BookingCache -.->|Shared Cache| Redis[(Redis)]
    CheckInCache -.->|Shared Cache| Redis
    
    style BookingCache fill:#fff9c4
    style CheckInCache fill:#fff9c4
```

---

## Database Schema

### Entity Relationship Diagram

```mermaid
erDiagram
    FLIGHTS ||--o{ SEATS : has
    FLIGHTS ||--o{ BOOKINGS : "has bookings for"
    BOOKINGS ||--o| CHECKINS : "can have"
    CHECKINS ||--|| SEATS : "assigns"
    
    FLIGHTS {
        uuid id PK
        string code
        string source
        string destination
        timestamp departure_time
        string flight_type
    }
    
    SEATS {
        uuid id PK
        uuid flight_id FK
        string seat_no
        string status
        boolean is_premium
        uuid booking_id FK "legacy"
        uuid checkin_id FK "new"
    }
    
    BOOKINGS {
        uuid id PK
        string pnr UK "6-char code"
        uuid flight_id FK
        string passenger_first_name
        string passenger_last_name
        string passport_number
        timestamp created_at
    }
    
    CHECKINS {
        uuid id PK
        uuid booking_id FK
        string seat_no
        decimal baggage_weight
        boolean baggage_fee_paid
        enum status "IN_PROGRESS|WAITING_PAYMENT|COMPLETED"
        timestamp checked_in_at
        timestamp completed_at
    }
```

---

## Data Flow

### Flight Booking System (Existing)

```mermaid
sequenceDiagram
    actor User
    participant UI as Booking UI
    participant API as Booking API :8080
    participant DB as PostgreSQL
    participant Redis
    
    User->>UI: Select date
    UI->>API: GET /flights?date=2026-02-16
    API->>DB: Query flights
    DB-->>API: Flight list
    API-->>UI: Flights
    
    User->>UI: Select flight
    UI->>API: GET /flights/:id/seats
    API->>DB: Query seats
    DB-->>API: Seat map
    API-->>UI: Seats
    
    User->>UI: Click seat 4A
    UI->>API: POST /seats/:id/hold
    API->>Redis: SETNX hold:flightID:4A EX 45
    Redis-->>API: Success
    API-->>UI: Held (45s timer)
    
    User->>UI: Submit details
    UI->>API: POST /bookings
    API->>DB: INSERT booking
    API->>DB: UPDATE seat (booking_id)
    API->>Redis: DEL hold:flightID:4A
    DB-->>API: Booking reference
    API-->>UI: Success
    UI-->>User: Booking confirmed
```

### Web Check-In System (New)

```mermaid
sequenceDiagram
    actor Passenger
    participant UI as Check-In UI
    participant API as Check-In API :8081
    participant DB as PostgreSQL
    participant Redis
    
    Passenger->>UI: Enter PNR + Last Name
    UI->>API: POST /webcheckin/lookup
    API->>DB: SELECT * FROM bookings WHERE pnr=? AND last_name=?
    DB-->>API: Booking found
    API->>DB: Check existing check-in
    DB-->>API: Not checked in
    API-->>UI: Flight details
    
    Passenger->>UI: Proceed to seats
    UI->>API: GET /webcheckin/:pnr/seats
    API->>DB: Get seat map (exclude checked-in seats)
    DB-->>API: Available seats
    API-->>UI: Seat map
    
    Passenger->>UI: Select seat 6B
    UI->>API: POST /webcheckin/:pnr/hold-seat
    API->>Redis: SETNX hold:PNR:6B EX 120
    Redis-->>API: Success
    API-->>UI: Held (120s timer)
    
    Passenger->>UI: Enter baggage: 30kg
    UI->>API: POST /webcheckin/:pnr/complete {30kg}
    
    alt Weight > 25kg
        API-->>UI: 402 Payment Required
        UI->>Passenger: Show payment modal
        Passenger->>UI: Simulate payment
        UI->>API: POST /webcheckin/:pnr/baggage-payment
        API->>DB: UPDATE check-in (fee_paid=true)
        DB-->>API: Success
        API-->>UI: Payment accepted
    end
    
    UI->>API: POST /webcheckin/:pnr/complete {30kg}
    API->>DB: INSERT INTO checkins
    API->>DB: UPDATE seat (checkin_id)
    API->>Redis: DEL hold:PNR:6B
    DB-->>API: Check-in complete
    API-->>UI: Boarding pass data
    UI-->>Passenger: Show boarding pass
```

---

## Infrastructure

### Docker Compose Configuration

```yaml
version: '3.8'

services:
  # Existing booking service
  booking-service:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
  
  # New check-in service
  checkin-service:
    build: ./backend_webcheckin
    ports:
      - "8081:8081"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
  
  # Frontend (serves both systems)
  frontend:
    build: ./frontend
    ports:
      - "5173:5173"
    depends_on:
      - booking-service
      - checkin-service
  
  # Shared PostgreSQL
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: skyhigh
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
  
  # Shared Redis
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

---

## Technology Stack

| Layer | Technology | Version | Purpose |
|-------|-----------|---------|---------|
| **Frontend** | React | 18.x | UI framework |
| | React Router | 6.x | Client-side routing |
| | TailwindCSS | 3.x | Styling |
| | Vite | 5.x | Build tool |
| **Backend** | Go | 1.21+ | Primary language |
| | Fiber | 2.x | HTTP framework |
| | GORM | 1.x | ORM |
| **Database** | PostgreSQL | 15.x | Relational data |
| | Redis | 7.x | Caching & seat holds |
| **DevOps** | Docker | 24.x | Containerization |
| | Docker Compose | 2.x | Orchestration |

---

## Deployment Architecture

```mermaid
graph TB
    subgraph "Production Environment"
        LB[Load Balancer<br/>NGINX]
        
        subgraph "Frontend Tier"
            FE1[Frontend Instance 1]
            FE2[Frontend Instance 2]
        end
        
        subgraph "Backend Tier"
            subgraph "Booking Service"
                BS1[Booking :8080 - Instance 1]
                BS2[Booking :8080 - Instance 2]
            end
            
            subgraph "Check-In Service"
                CS1[Check-In :8081 - Instance 1]
                CS2[Check-In :8081 - Instance 2]
            end
        end
        
        subgraph "Data Tier"
            PG_Primary[(PostgreSQL<br/>Primary)]
            PG_Replica[(PostgreSQL<br/>Read Replica)]
            Redis_Cluster[(Redis<br/>Cluster)]
        end
        
        LB --> FE1
        LB --> FE2
        
        FE1 --> BS1
        FE1 --> BS2
        FE1 --> CS1
        FE1 --> CS2
        
        FE2 --> BS1
        FE2 --> BS2
        FE2 --> CS1
        FE2 --> CS2
        
        BS1 --> PG_Primary
        BS2 --> PG_Replica
        CS1 --> PG_Primary
        CS2 --> PG_Replica
        
        BS1 --> Redis_Cluster
        BS2 --> Redis_Cluster
        CS1 --> Redis_Cluster
        CS2 --> Redis_Cluster
    end
    
    style LB fill:#b3e5fc
    style PG_Primary fill:#ffccbc
    style Redis_Cluster fill:#fff9c4
```

---

## Security Considerations

| Concern | Mitigation |
|---------|------------|
| **PNR Enumeration** | Rate limiting (5 attempts/minute/IP) |
| **SQL Injection** | GORM parameterized queries |
| **XSS** | React auto-escaping, CSP headers |
| **CORS** | Whitelist frontend origin only |
| **Seat Hold Abuse** | Redis TTL enforcement, IP-based limits |
| **Payment Fraud** | Simulated for MVP, integrate payment gateway for production |

---

## Monitoring & Observability

```mermaid
graph LR
    subgraph "Application Metrics"
        Prometheus[Prometheus]
        Grafana[Grafana Dashboards]
    end
    
    subgraph "Logging"
        Logs[Application Logs]
        ELK[ELK Stack<br/>Elasticsearch + Kibana]
    end
    
    subgraph "Tracing"
        Jaeger[Jaeger<br/>Distributed Tracing]
    end
    
    BookingService[Booking Service] --> Prometheus
    CheckInService[Check-In Service] --> Prometheus
    Prometheus --> Grafana
    
    BookingService --> Logs
    CheckInService --> Logs
    Logs --> ELK
    
    BookingService --> Jaeger
    CheckInService --> Jaeger
    
    style Prometheus fill:#ffe0b2
    style Grafana fill:#c8e6c9
```

**Key Metrics to Monitor:**
- Request latency (P50, P95, P99)
- Seat hold expiry rate
- Check-in completion rate
- Database connection pool usage
- Redis memory usage
- Error rates by endpoint

---

## Scalability Considerations

### Horizontal Scaling
- **Frontend**: Stateless, scale to N instances behind load balancer
- **Backend Services**: Stateless, scale independently based on load
- **Database**: Read replicas for GET requests, primary for writes
- **Redis**: Cluster mode for high availability

### Performance Optimization
- **Seat Map Caching**: Cache seat maps in Redis (1-minute TTL)
- **Database Indexing**: 
  - `bookings(pnr)` - For fast PNR lookup
  - `checkins(booking_id)` - For check-in status queries
  - `seats(flight_id, status)` - For seat availability
- **Connection Pooling**: Max 50 connections per service instance

---

## Disaster Recovery

| Scenario | Recovery Strategy | RTO | RPO |
|----------|------------------|-----|-----|
| Database Failure | Failover to replica | 30s | 0 (synchronous replication) |
| Redis Failure | Rebuild holds from DB | 1min | 0 (seats in DB) |
| Service Crash | Auto-restart (Docker) | 10s | 0 |
| Data Center Outage | Multi-region deployment | 5min | 1min |

---

## Future Enhancements

1. **Real Payment Integration**: Stripe/PayPal for baggage fees
2. **Mobile Apps**: React Native iOS/Android apps
3. **Boarding Pass PDF**: Generate downloadable PDFs
4. **Email Notifications**: Send boarding pass via email
5. **SMS Reminders**: Check-in reminders 24h before flight
6. **Kiosk Mode**: Dedicated UI for airport kiosks
7. **Admin Dashboard**: Monitor check-ins, override holds

---

**Document Version**: 2.0  
**Last Review**: 2026-02-16  
**Next Review**: 2026-03-01
