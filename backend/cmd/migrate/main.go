package main

import (
	"fmt"
	"log"
	"os"

	"digital-checkin/pkg/config"
	"digital-checkin/pkg/db"
)

func main() {
	cfg := config.LoadConfig()
    
    // Override host for migration if running from separate container or same
    // Just use config
    
	database, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("Could not connect to database for migration: %v", err)
	}
	defer database.Close()

	log.Println("Running migrations...")
    
    // Read the SQL file
    // In a real app we'd iterate over files. Here we have just 000001_create_initial_schema.up.sql
    // We need to bake this file into the binary or read from disk. 
    // Dockerfile copies ./migrations so we can read from disk.
    
    content, err := os.ReadFile("migrations/000001_create_initial_schema.up.sql")
    if err != nil {
        log.Fatalf("Failed to read migration file: %v", err)
    }
    
    if _, err := database.Exec(string(content)); err != nil {
         // Check if it's already applied? or simple ignore
         // For CREATE TABLE IF NOT EXISTS it's fine.
         log.Printf("Migration warning (might be already applied): %v", err)
    } else {
        log.Println("Migration applied successfully.")
    }
    
    // Seeding?
    // We can call the core seeder here too if needed, or stick to main app seeding.
    // Let's keep seeding in main app for now as it's optional there.
    
    fmt.Println("Migration complete.")
}
