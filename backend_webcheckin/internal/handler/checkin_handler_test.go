package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"backend_webcheckin/internal/handler"
	"backend_webcheckin/internal/repository"
	"backend_webcheckin/internal/service"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestApp(t *testing.T) (*fiber.App, sqlmock.Sqlmock, redismock.ClientMock) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)

	rdb, mockRedis := redismock.NewClientMock()

	repo := &repository.Repository{
		DB:    gormDB,
		Redis: rdb,
	}
	svc := service.NewCheckInService(repo, 300)
	h := handler.NewCheckInHandler(svc)

	app := fiber.New()
	app.Post("/api/webcheckin/lookup", h.LookupBooking)
	app.Get("/api/webcheckin/:pnr/seats", h.GetSeats)
	app.Post("/api/webcheckin/:pnr/hold-seat", h.HoldSeat)
	app.Post("/api/webcheckin/:pnr/complete", h.CompleteCheckIn)
	app.Post("/api/webcheckin/:pnr/baggage-payment", h.ProcessBaggagePayment)
	app.Post("/api/webcheckin/admin/seed-bookings", h.SeedBookings)

	return app, mockDB, mockRedis
}

func TestLookupBooking_Success(t *testing.T) {
	app, mockDB, _ := setupTestApp(t)

	bookingID := uuid.New()
	flightID := uuid.New()

	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"webcheckin_bookings\"").
		WillReturnRows(sqlmock.NewRows([]string{"id", "pnr", "flight_id", "passenger_first_name", "passenger_last_name"}).
			AddRow(bookingID, "ABC123", flightID, "John", "Doe"))

	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"flights\" WHERE \"flights\"\\.\"id\" = \\$1").
		WithArgs(flightID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(flightID, "FL123"))

	mockDB.ExpectQuery("(?i)SELECT s.seat_no FROM bookings b JOIN seats s").
		WillReturnRows(sqlmock.NewRows([]string{"seat_no"}).AddRow("1A"))

	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"checkins\" WHERE booking_id = \\$1").
		WithArgs(bookingID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "booking_id", "status"}).AddRow(uuid.New(), bookingID, "COMPLETED"))

	reqBody := `{"pnr": "ABC123", "lastName": "Doe"}`
	req := httptest.NewRequest("POST", "/api/webcheckin/lookup", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var resBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.Equal(t, "COMPLETED", resBody["checkinStatus"])
}

func TestGetSeats_Success(t *testing.T) {
	app, mockDB, _ := setupTestApp(t)

	flightID := uuid.New()

	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"seats\" WHERE flight_id = \\$1 ORDER BY seat_no ASC").
		WithArgs(flightID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "flight_id", "seat_no", "is_booked"}).
			AddRow(uuid.New(), flightID, "1A", false))

	req := httptest.NewRequest("GET", "/api/webcheckin/ABC123/seats?flightId="+flightID.String(), nil)

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var resBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&resBody)
	assert.Equal(t, flightID.String(), resBody["flightId"])
}

func TestHoldSeat_Success(t *testing.T) {
	app, mockDB, mockRedis := setupTestApp(t)

	flightID := uuid.New()
	seatNo := "1A"

	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"seats\" WHERE flight_id = \\$1").
		WithArgs(flightID.String(), seatNo).
		WillReturnRows(sqlmock.NewRows([]string{"id", "seat_no", "is_booked"}).
			AddRow(uuid.New(), seatNo, false))

	mockRedis.ExpectSetNX("checkin:hold:ABC123:1A", flightID.String(), 300*time.Second).SetVal(true)

	reqBody := `{"flightId": "` + flightID.String() + `", "seatNo": "1A"}`
	req := httptest.NewRequest("POST", "/api/webcheckin/ABC123/hold-seat", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestCompleteCheckIn_Success(t *testing.T) {
	app, mockDB, mockRedis := setupTestApp(t)

	flightID := uuid.New()
	bookingID := uuid.New()
	seatID := uuid.New()
	seatNo := "1A"

    // LookupBooking queries
	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"webcheckin_bookings\"").
		WillReturnRows(sqlmock.NewRows([]string{"id", "pnr", "flight_id"}).AddRow(bookingID, "ABC123", flightID))

	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"flights\"").
		WillReturnRows(sqlmock.NewRows([]string{"id", "code", "departure_time"}).AddRow(flightID, "FL123", time.Now().Add(24*time.Hour)))

	mockDB.ExpectQuery("(?i)SELECT s.seat_no FROM bookings b").
		WillReturnRows(sqlmock.NewRows([]string{"seat_no"}))

	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"checkins\"").
		WillReturnError(gorm.ErrRecordNotFound) // Not checked in yet

    // Complete check-in queries
	mockDB.ExpectQuery("(?i)SELECT \\* FROM \"seats\"").
		WithArgs(flightID.String(), seatNo).
		WillReturnRows(sqlmock.NewRows([]string{"id", "seat_no", "is_booked"}).AddRow(seatID, seatNo, false))
        
    mockDB.ExpectBegin()
	mockDB.ExpectQuery("(?i)INSERT INTO \"checkins\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
    mockDB.ExpectCommit()

    // Since Update occurs via Model(&models.Seat{}).Where("id = ?", seatID).Updates(...)
    mockDB.ExpectBegin()
	mockDB.ExpectExec("(?i)UPDATE \"seats\" SET").
		WillReturnResult(sqlmock.NewResult(1, 1))
    mockDB.ExpectCommit()

	mockRedis.ExpectDel("checkin:hold:ABC123:1A").SetVal(1)

    // And also potentially updating bookings, but initial seat is empty here.

	reqBody := `{"seatNo": "1A", "baggageWeight": 20.0}`
	req := httptest.NewRequest("POST", "/api/webcheckin/ABC123/complete", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestProcessBaggagePayment_Success(t *testing.T) {
	app, mockDB, _ := setupTestApp(t)

	checkInID := uuid.New()

    mockDB.ExpectBegin()
	mockDB.ExpectExec("(?i)UPDATE \"checkins\" SET \"baggage_fee_amount\"=\\$1,\"baggage_fee_paid\"=\\$2").
		WithArgs(100.0, true, checkInID.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
    mockDB.ExpectCommit()

	reqBody := `{"checkInId": "` + checkInID.String() + `", "feeAmount": 100.0, "paymentMethod": "CARD"}`
	req := httptest.NewRequest("POST", "/api/webcheckin/ABC123/baggage-payment", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
