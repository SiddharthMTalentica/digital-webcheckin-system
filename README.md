# ✈️ SkyHigh Airlines - Digital Check-In System

Industrial-level dual-system web application featuring flight booking and web check-in services.

## 🎯 Features

### Flight Booking System (`/flight-booking-system`)
- Search flights by date (60-day range, 180 flights)
- Interactive seat map with real-time availability
- 45-second seat hold timer
- Passenger details form
- Booking confirmation

### Web Check-In System (`/web-check-in`) ⭐ NEW
- PNR lookup with last name verification
- 120-second seat hold timer
- Baggage weight management (25kg free allowance)
- Automatic excess baggage fee calculation ($10/kg)
- Simulated payment flow for overweight baggage
- Digital boarding pass generation
- QR code and barcode simulation

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                 Frontend (React + Vite)                  │
│            http://localhost:5173                         │
│  ┌──────────┐  ┌─────────────┐  ┌─────────────────┐   │
│  │ Homepage │→ │   Booking   │  │  Web Check-In   │   │
│  └──────────┘  │   Flow      │  │     Flow        │   │
│                 └─────────────┘  └─────────────────┘   │
└─────────────────────────────────────────────────────────┘
                      ↓                    ↓
         ┌────────────────────┐  ┌────────────────────┐
         │  Booking Service   │  │  CheckIn Service   │
         │  :8080 (Go)        │  │  :8081 (Go)        │
         └────────────────────┘  └────────────────────┘
                      ↓                    ↓
         ┌────────────────────────────────────────────┐
         │     PostgreSQL + Redis                      │
         │     (Shared Database)                       │
         └────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)

### Start All Services

```bash
./start.sh
```

This will start:
- **PostgreSQL** (port 5432)
- **Redis** (port 6379)
- **Booking Backend** (port 8080)
- **CheckIn Backend** (port 8081)
- **Frontend** (port 5173)

Then visit: **http://localhost:5173**

## 📝 Testing Web Check-In

### Test PNRs (automatically seeded)
- **PNR**: `ABC123`, **Last Name**: `Doe`
- **PNR**: `XYZ789`, **Last Name**: `Smith`

### Test Scenarios

**Normal Check-In**
1. Enter PNR: `ABC123`, Last Name: `Doe`
2. Select any green (available) seat
3. Enter baggage weight: `20` kg
4. ✅ Check-in completes → Boarding pass displayed

**Excess Baggage Payment**
1. Enter PNR: `XYZ789`, Last Name: `Smith`
2. Select a seat
3. Enter baggage weight: `30` kg
4. 💳 Payment modal appears ($50 fee)
5. Click "Pay & Complete"
6. ✅ Boarding pass displayed

**Seat Hold Timer Expiry**
1. Select a seat
2. Wait 120 seconds without submitting baggage
3. ⏱️ Seat hold expires, must reselect

## 🛠️ Tech Stack

- **Backend**: Go, Fiber, GORM, Redis
- **Frontend**: React, React Router, Tailwind CSS
- **Database**: PostgreSQL, Redis
- **Deployment**: Docker, Docker Compose

## 📂 Project Structure

```
.
├── backend/                    # Booking service (port 8080)
├── backend_webcheckin/         # Check-in service (port 8081)
│   ├── cmd/main.go
│   ├── internal/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── service/
│   │   └── handler/
│   └── migrations/
├── frontend/
│   └── src/
│       ├── components/
│       │   ├── Homepage.jsx
│       │   ├── SeatMap.jsx
│       │   └── webcheckin/
│       │       ├── PNRLookup.jsx
│       │       ├── WebCheckInApp.jsx
│       │       ├── PaymentModal.jsx
│       │       └── BoardingPass.jsx
│       └── main.jsx
├── docker-compose.yml
├── start.sh
├── tasks.md                    # Booking system tasks
└── webcheckin_tasks.md         # Check-in system tasks
```

## 🔌 API Endpoints

### Booking Service (8080)
- `GET /api/flights?date=YYYY-MM-DD`
- `GET /api/flights/:id/seats`
- `POST /api/flights/:id/seats/:seat/hold`
- `POST /api/bookings`

### Check-In Service (8081)
- `POST /api/webcheckin/lookup` - PNR verification
- `GET /api/webcheckin/:pnr/seats` - Seat availability
- `POST /api/webcheckin/:pnr/hold-seat` - Hold seat (120s)
- `POST /api/webcheckin/:pnr/complete` - Complete check-in
- `POST /api/webcheckin/:pnr/baggage-payment` - Process payment

## 🎨 Design Highlights

- **Industrial-level UI**: Gradient backgrounds, smooth animations
- **Professional layouts**: Grid-based, responsive design
- **Visual feedback**: Progress indicators, loading states, error messages
- **Micro-interactions**: Hover effects, color transitions
- **Accessibility**: Proper labels, keyboard navigation

## 📊 Key Metrics

- **Backend Services**: 2 microservices
- **API Endpoints**: 11 total (5 booking + 6 check-in)
- **Frontend Components**: 10+
- **Database Tables**: 6 (flights, seats, bookings_old, bookings, checkins, passengers)
- **Total Lines of Code**: ~3,000+

## 🔄 Development Workflow

### Run Backend Only (Local)
```bash
cd backend_webcheckin
go run cmd/main.go
```

### Run Frontend Only (Local)
```bash
cd frontend
npm install
npm run dev
```

### Rebuild Docker Images
```bash
docker-compose up --build
```

### Stop All Services
```bash
docker-compose down
```

## 📖 Documentation

- [Product Requirements Document](/PRD.md)
- [System Architecture](/ARCHITECTURE.md)
- [Workflow Design](/WORKFLOW_DESIGN.md)
- [API Specification](/API-SPECIFICATION.md)
- [Project Structure](/PROJECT_STRUCTURE.md)

## ✅ Requirements Compliance

| Requirement | Status |
|------------|--------|
| PNR + Last Name lookup | ✅ |
| 120-second seat hold | ✅ |
| 25kg baggage limit | ✅ |
| Excess fee calculation | ✅ |
| Payment flow simulation | ✅ |
| Boarding pass generation | ✅ |
| Concurrent safety (Redis locks) | ✅ |

## 🚧 Future Enhancements

- [ ] Real payment gateway (Stripe/PayPal)
- [ ] Email/SMS notifications
- [ ] PDF boarding pass download
- [ ] Admin dashboard
- [ ] Analytics & reporting
- [ ] Load testing

---

**Built with ❤️ for SkyHigh Airlines**
