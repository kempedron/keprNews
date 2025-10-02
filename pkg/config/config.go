package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret       string
	AuthServisePort string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}
	return &Config{
		JWTSecret:       GetEnv("JWT_SECRET", ""),
		AuthServisePort: GetEnv("AUTH_SERVICE_PORT", "8081"),
	}
}

func GetEnv(key string, defaultString string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultString
}
