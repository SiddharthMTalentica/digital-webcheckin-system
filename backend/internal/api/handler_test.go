package api_test

import (
	"bytes"
	"digital-checkin/internal/api"
	"digital-checkin/internal/repository"
	"digital-checkin/internal/service"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestApp(t *testing.T) (*api.Handler, sqlmock.Sqlmock, redismock.ClientMock) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)

	repo := repository.NewRepository(db)

	rdb, mockRedis := redismock.NewClientMock()
	seatSvc := service.NewSeatService(repo, rdb)

	handler := api.NewHandler(seatSvc)

	return handler, mockDB, mockRedis
}

func TestGetFlights(t *testing.T) {
	handler, mockDB, _ := setupTestApp(t)

	now := time.Now()
	flightID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "code", "source", "destination", "departure_time", "plane_type", "created_at", "updated_at"}).
		AddRow(flightID, "FL123", "NYC", "LAX", now, "A320", now, now)

	mockDB.ExpectQuery("(?i)SELECT .* FROM flights").WillReturnRows(rows)

	req, _ := http.NewRequest("GET", "/api/v1/flights", nil)
	rr := httptest.NewRecorder()

	router := handler.Routes()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var flights []map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &flights)
	assert.NoError(t, err)
	assert.Len(t, flights, 1)
	assert.Equal(t, "FL123", flights[0]["code"])
}

func TestGetSeats_Success(t *testing.T) {
	handler, mockDB, mockRedis := setupTestApp(t)

	flightID := uuid.New()
	now := time.Now()
	seatID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "flight_id", "seat_no", "row_num", "col_num", "category", "is_booked", "created_at", "updated_at"}).
		AddRow(seatID, flightID, "1A", 1, 1, "Economy", false, now, now)

	mockDB.ExpectQuery("(?i)SELECT .* FROM seats WHERE flight_id").
		WithArgs(flightID).
		WillReturnRows(rows)

	key := "hold:" + flightID.String() + ":1A"
	mockRedis.ExpectExists(key).SetVal(0) // Not held

	req, _ := http.NewRequest("GET", "/api/v1/flights/"+flightID.String()+"/seats", nil)
	rr := httptest.NewRecorder()

	router := handler.Routes()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, flightID.String(), resp["flight_id"])

	seatsRaw, ok := resp["seats"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, seatsRaw, 1)

	seatMap := seatsRaw[0].(map[string]interface{})
	assert.Equal(t, "1A", seatMap["seat_no"])
	assert.Equal(t, "AVAILABLE", seatMap["status"])
}

func TestHoldSeat_Success(t *testing.T) {
	handler, mockDB, mockRedis := setupTestApp(t)

	flightID := uuid.New()
	seatNo := "1A"
	userID := "user123"

	mockDB.ExpectQuery("(?i)SELECT is_booked FROM seats WHERE flight_id = \\$1 AND seat_no = \\$2").
		WithArgs(flightID, seatNo).
		WillReturnRows(sqlmock.NewRows([]string{"is_booked"}).AddRow(false))

	key := "hold:" + flightID.String() + ":" + seatNo
	mockRedis.ExpectSetNX(key, userID, 45*time.Second).SetVal(true)

	reqBody := `{"user_id": "` + userID + `"}`
	req, _ := http.NewRequest("POST", "/api/v1/flights/"+flightID.String()+"/seats/"+seatNo+"/hold", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := handler.Routes()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Seat held successfully", resp["message"])
	assert.Equal(t, seatNo, resp["seat_no"])
}

func TestConfirmCheckIn_Success(t *testing.T) {
	handler, mockDB, mockRedis := setupTestApp(t)

	flightID := uuid.New()
	seatNo := "1A"
	userID := "user123"
	seatID := uuid.New()

	key := "hold:" + flightID.String() + ":" + seatNo
	mockRedis.ExpectGet(key).SetVal(userID)
	mockRedis.ExpectDel(key).SetVal(1)

	mockDB.ExpectBegin()
	mockDB.ExpectQuery("(?i)SELECT id, is_booked FROM seats WHERE flight_id = \\$1 AND seat_no = \\$2 FOR UPDATE").
		WithArgs(flightID, seatNo).
		WillReturnRows(sqlmock.NewRows([]string{"id", "is_booked"}).AddRow(seatID, false))

	mockDB.ExpectExec("(?i)INSERT INTO bookings").
		WithArgs(sqlmock.AnyArg(), flightID, seatID, "John", "Doe", "AB12345", "CONFIRMED", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockDB.ExpectExec("(?i)UPDATE seats SET is_booked = TRUE").
		WithArgs(seatID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mockDB.ExpectCommit()

	reqBody := `{
        "flight_id": "` + flightID.String() + `",
        "seat_no": "1A",
        "user_id": "` + userID + `",
        "first_name": "John",
        "last_name": "Doe",
        "passport": "AB12345",
        "baggage_weight": 20.0
    }`
	req, _ := http.NewRequest("POST", "/api/v1/checkin/confirm", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router := handler.Routes()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "John", resp["passenger_first_name"])
	assert.Equal(t, "CONFIRMED", resp["status"])
}
