package api

import (
	"digital-checkin/internal/service"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
)

type Handler struct {
	SeatService *service.SeatService
}

func NewHandler(seatService *service.SeatService) *Handler {
	return &Handler{SeatService: seatService}
}

func (h *Handler) GetFlights(w http.ResponseWriter, r *http.Request) {
	flights, err := h.SeatService.GetAllFlights(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch flights", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flights)
}

func (h *Handler) GetSeats(w http.ResponseWriter, r *http.Request) {
	flightIDStr := chi.URLParam(r, "flightID")
	flightID, err := uuid.Parse(flightIDStr)
	if err != nil {
		http.Error(w, "Invalid flight ID", http.StatusBadRequest)
		return
	}

	seats, err := h.SeatService.GetFlightSeats(r.Context(), flightID)
	if err != nil {
		http.Error(w, "Failed to fetch seats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"flight_id": flightID,
		"seats":     seats,
	})
}

type HoldSeatRequest struct {
    UserID string `json:"user_id"`
}

func (h *Handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
    flightIDStr := chi.URLParam(r, "flightID")
    seatNo := chi.URLParam(r, "seatNo")
    
    flightID, err := uuid.Parse(flightIDStr)
    if err != nil {
        http.Error(w, "Invalid flight ID", http.StatusBadRequest)
        return
    }

    var req HoldSeatRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.UserID == "" {
        http.Error(w, "user_id is required", http.StatusBadRequest)
        return
    }

    token, expiresAt, err := h.SeatService.HoldSeat(r.Context(), flightID, seatNo, req.UserID)
    if err != nil {
        if err.Error() == fmt.Sprintf("seat %s is currently held by another user", seatNo) || 
           err.Error() == fmt.Sprintf("seat %s is already booked", seatNo) {
            http.Error(w, err.Error(), http.StatusConflict)
            return   
        }
        if err.Error() == "seat not found" {
             http.Error(w, err.Error(), http.StatusNotFound)
             return
        }
        
        http.Error(w, "Failed to hold seat", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Seat held successfully",
        "seat_no": seatNo,
        "token": token, 
        "expires_at": expiresAt,
    })
}

// ConfirmCheckInRequest payload
type ConfirmCheckInRequest struct {
    FlightID      string  `json:"flight_id"`
    SeatNo        string  `json:"seat_no"`
    UserID        string  `json:"user_id"`
    FirstName     string  `json:"first_name"`
    LastName      string  `json:"last_name"`
    Passport      string  `json:"passport"`
    BaggageWeight float64 `json:"baggage_weight"`
}

func (h *Handler) ConfirmCheckIn(w http.ResponseWriter, r *http.Request) {
    var req ConfirmCheckInRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    flightID, err := uuid.Parse(req.FlightID)
    if err != nil {
        http.Error(w, "Invalid flight ID", http.StatusBadRequest)
        return
    }
    
    booking, err := h.SeatService.ConfirmCheckIn(r.Context(), service.CheckInRequest{
        FlightID:           flightID,
        SeatNo:             req.SeatNo,
        UserID:             req.UserID,
        PassengerFirstName: req.FirstName,
        PassengerLastName:  req.LastName,
        PassengerPassport:  req.Passport,
        BaggageWeight:      req.BaggageWeight,
    })
    
    if err != nil {
        if err.Error() == "seat hold expired or invalid" || err.Error() == "seat is held by another user" {
             http.Error(w, err.Error(), http.StatusConflict) // 409
             return
        }
        if err.Error() == "baggage weight exceeds 25kg limit. Payment required" {
             http.Error(w, err.Error(), http.StatusPaymentRequired) // 402
             return
        }
        http.Error(w, fmt.Sprintf("Check-in failed: %v", err), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(booking)
}


func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
    
    // Basic CORS
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"https://*", "http://*"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: true,
        MaxAge:           300,
    }))

    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/flights", h.GetFlights)
        r.Get("/flights/{flightID}/seats", h.GetSeats)
        r.Post("/flights/{flightID}/seats/{seatNo}/hold", h.HoldSeat)
        r.Post("/checkin/confirm", h.ConfirmCheckIn)
    })
	return r
}
