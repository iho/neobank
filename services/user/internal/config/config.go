package config

import (
	"os"
)

type Config struct {
	DatabaseURL     string
	HTTPPort        string
	GRPCPort        string
	LedgerAddr      string
	RedisURL        string
	JWTSecret       string
	KafkaBrokers    string
	NotificationURL string
}

func Load() Config {
	return Config{
		DatabaseURL:     env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:        env("HTTP_PORT", "8081"),
		GRPCPort:        env("GRPC_PORT", "50052"),
		LedgerAddr:      env("LEDGER_GRPC_ADDR", "localhost:50051"),
		RedisURL:        env("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:       env("JWT_SECRET", "dev-secret-change-me"),
		KafkaBrokers:    env("KAFKA_BROKERS", ""),
		NotificationURL: env("NOTIFICATION_SERVICE_URL", "http://localhost:8083/api/v1/internal/events"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}