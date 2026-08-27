// config/database.go
package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	FrontendURL   string
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
	FrontendURL := os.Getenv("FRONTEND_URL")
	if FrontendURL == "" {
		FrontendURL = "https://nid.xyz"
	}
	return &Config{
		DatabaseURL: dbURL,
		Port:        port,
		FrontendURL: FrontendURL,
	}
}
