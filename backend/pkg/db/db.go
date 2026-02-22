package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
)

func Connect(dbUrl string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Connected to PostgreSQL successfully")
	return db, nil
}
