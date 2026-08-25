package config

import (
	"fmt"
	"os"
	"strings"
)


func GetEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return fallback
	}

	return value
}

func GetRequiredEnv(key string) (string, error) {
	value := strings.TrimSpace(os.Getenv(key))

	if value == "" {
		return "", fmt.Errorf("%s environment variable is required", key)
	}

	return value, nil
}
