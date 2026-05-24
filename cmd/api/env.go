package main

import (
	"fmt"
	"os"
)

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getDatabaseDSN() string {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}

	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "istore"),
		getEnv("DB_PASSWORD", "istore"),
		getEnv("DB_NAME", "istore"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_SSLMODE", "disable"),
		getEnv("DB_TIMEZONE", "America/Sao_Paulo"),
	)
}
