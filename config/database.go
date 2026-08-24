// config/database.go
package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
}

func LoadConfig() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/nid_db?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
	}
}
