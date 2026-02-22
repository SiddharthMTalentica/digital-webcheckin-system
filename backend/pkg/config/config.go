package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl      string
	RedisAddr  string
	ServerPort string
}

func LoadConfig() *Config {
	// Attempt to load .env, but don't fail if it doesn't exist (e.g. in Docker)
	_ = godotenv.Load()

	return &Config{
		DBUrl:      getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/skyhigh_db?sslmode=disable"),
		RedisAddr:  getEnv("REDIS_ADDR", "localhost:6379"),
		ServerPort: getEnv("SERVER_PORT", ":8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
