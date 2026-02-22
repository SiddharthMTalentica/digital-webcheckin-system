# Product Requirements Document (PRD) - SkyHigh Core

## 1. Problem Description
SkyHigh Airlines faces significant challenges during peak check-in windows. Hundreds of passengers attempt to select seats, add baggage, and complete check-in simultaneously. The existing systems struggle with:
- **Seat Conflicts**: Double bookings due to race conditions.
- **Scalability**: Performance degradation under high load.
- **User Experience**: Frustrations caused by failed transactions or slow updates.

The goal is to build **SkyHigh Core**, a robust backend service that guarantees a fast, safe, and automated digital check-in experience.

*Note: The system has been architected as a Dual-System Backend:*
- ***Booking Service (Port 8080)***: *Handles initial flight searches and reservations.*
- ***Web Check-In Service (Port 8081)***: *Handles high-traffic, time-sensitive check-in tasks (seat locks, baggage fees).*

## 2. Goals & Objectives
- **Conflict-Free Assignments**: Guarantee that no two passengers can book the same seat (Hard Guarantee).
- **Time-Bound Reservations**: Implement a strict 120-second hold on selected seats to allow for completion of check-in details.
- **High Performance**: Seat map data must verify P95 latency < 1 second.
- **Baggage Handling**: Enforce weight limits (25kg) and trigger payment flows for excess baggage.

## 3. Key Users
- **Passengers**: End-users performing self-check-in via web or kiosk.
- **Airline Staff**: Admin users monitoring flight status and seat maps (future scope).

## 4. Non-Functional Requirements (NFRs)
- **Concurrency**: Must handle hundreds of concurrent requests without data corruption.
- **Availability**: System should remain available during high traffic.
- **Latency**: API response times should be minimal (< 200ms for holds, < 1s for maps).
- **Scalability**: Architecture should support horizontal scaling (stateless API, distributed locking).
- **Consistency**: Strong consistency for seat reservation state.
