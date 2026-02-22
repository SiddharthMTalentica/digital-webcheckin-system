# Chat History / Design Journey

## Summary of Collaboration

The design and implementation of SkyHigh Core was a collaborative process between the User (Architect) and the AI Assistant.

### Phase 1: Requirements Analysis
- **Input**: User provided the `SkyHigh Core – Digital Check-In System I1-I2.pdf`.
- **Action**: AI extracted core requirements:
    1.  Seat States: Available -> Held -> Confirmed.
    2.  120-second atomic hold.
    3.  Baggage validation logic.
    4.  High-performance read requirements.

### Phase 2: Implementation Planning
- **Decision**: We adopted a **Phased Implementation Approach** to ensure structured delivery.
    - Phase 1: Foundation (Go module, Docker).
    - Phase 2: Core Domain (DB Schema, Seat Map).
    - Phase 3: Concurrency (Redis Locking).
    - Phase 4: Business Rules (Check-in, Baggage).
    - Phase 5: Documentation & Testing.

### Phase 3: Key Technical Decisions (Architecture)
- **Concurrency**: We debated between DB row locking vs. Redis. We chose **Redis (Redlock/SET NX)** for the "hold" phase because it perfectly matches the ephemeral, high-throughput nature of the requirement (120s TTL). DB locking would be too heavy for just "selecting" a seat.
- **Database Schema**: Normalized schema (`flights`, `seats`, `bookings`) was chosen to ensure data integrity for confirmed bookings. `seats` table has `is_booked` flag for fast availability checks.
- **Project Structure**: Standard Go layout (`cmd`, `internal`, `pkg`) was selected for maintainability and clear separation of concerns.

### Phase 4: Verification
- We verified the implementation by:
    1.  Compiling the backend successfully.
    2.  Reviewing the code against the "Hard Guarantee" requirement (Atomic Redis operations confirmed).
    3.  Ensuring all NFRs (Latency, Scalability) are addressed by the chosen architecture.
