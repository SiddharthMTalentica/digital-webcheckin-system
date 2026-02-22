package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"backend_webcheckin/internal/handler"
	"backend_webcheckin/internal/repository"
	"backend_webcheckin/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Configuration
	port := getEnv("PORT", "8081")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvAsInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "admin")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "skyhigh")
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnvAsInt("REDIS_PORT", 6379)
	redisPassword := getEnv("REDIS_PASSWORD", "")
	holdDuration := getEnvAsInt("HOLD_DURATION", 120)

	// Initialize repository
	repo, err := repository.NewRepository(
		dbHost, dbUser, dbPassword, dbName, dbPort,
		redisHost, redisPort, redisPassword,
	)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer repo.Close()

	log.Println("✓ Connected to PostgreSQL and Redis")

	// Run migrations (simple approach)
	if err := runMigrations(repo); err != nil {
		log.Printf("Warning: Migration error: %v", err)
	}

	// Initialize service
	checkInService := service.NewCheckInService(repo, holdDuration)

	// Seed test bookings (ABC123, XYZ789) if they don't exist
	if err := checkInService.SeedTestBookings(context.Background()); err != nil {
		log.Printf("Warning: Failed to seed test bookings: %v", err)
	} else {
		log.Println("✓ Test bookings seeded (ABC123, XYZ789)")
	}

	// Initialize handler
	checkInHandler := handler.NewCheckInHandler(checkInService)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "SkyHigh Web Check-In Service",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} (${latency})\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "webcheckin",
			"time":    time.Now(),
		})
	})

	// API Routes
	api := app.Group("/api/webcheckin")

	// Check-in routes
	api.Post("/lookup", checkInHandler.LookupBooking)
	api.Get("/:pnr/seats", checkInHandler.GetSeats)
	api.Post("/:pnr/hold-seat", checkInHandler.HoldSeat)
	api.Post("/:pnr/complete", checkInHandler.CompleteCheckIn)
	api.Post("/:pnr/baggage-payment", checkInHandler.ProcessBaggagePayment)

	// Admin routes (development only)
	admin := app.Group("/api/webcheckin/admin")
	admin.Post("/seed-bookings", checkInHandler.SeedBookings)

	// Start server
	log.Printf("🚀 Web Check-In Service starting on port %s", port)
	log.Printf("📝 Hold Duration: %d seconds", holdDuration)
	log.Fatal(app.Listen(":" + port))
}

func runMigrations(repo *repository.Repository) error {
	// Read and execute migration file
	migrationSQL, err := os.ReadFile("migrations/001_create_checkin_tables.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if err := repo.DB.Exec(string(migrationSQL)).Error; err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	log.Println("✓ Migrations executed successfully")
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
