package config

import (
	"os"
)

type Config struct {
	DatabaseURL        string
	JWTSecret          string
	Port               string
	DBMaxConns         string
	DBMinConns         string
	DBMaxConnIdleTime  string
}

func Load() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:5433/prompt_management?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "super-secret-jwt-key"),
		Port:               getEnv("PORT", "8080"),
		DBMaxConns:        getEnv("DB_MAX_CONNS", "25"),
		DBMinConns:        getEnv("DB_MIN_CONNS", "2"),
		DBMaxConnIdleTime: getEnv("DB_MAX_CONN_IDLE_TIME", "5m"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
