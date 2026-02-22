package main

import (
	"log"
	"net/http"

	"digital-checkin/internal/api"
	"digital-checkin/internal/core"
	"digital-checkin/internal/repository"
	"digital-checkin/internal/service"
	"digital-checkin/pkg/config"
	"digital-checkin/pkg/db"
	"digital-checkin/pkg/redis"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Connect to Database
	database, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	// 3. Connect to Redis
	rdb, err := redis.Connect(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Could not connect to redis: %v", err)
	}
	defer rdb.Close()

	// 4. Run Seeder (Optional, usually behind a flag or check)
	seeder := core.NewSeeder(database)
	if err := seeder.Seed(); err != nil {
		log.Printf("Data seeding warning: %v", err)
	}

	// 5. Initialize Layers
	repo := repository.NewRepository(database)
	seatService := service.NewSeatService(repo, rdb)
	handler := api.NewHandler(seatService)

	// 6. Start Server
	srv := &http.Server{
		Addr:    cfg.ServerPort,
		Handler: handler.Routes(),
	}

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
