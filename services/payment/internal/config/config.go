package config

import "os"

type Config struct {
	DatabaseURL     string
	HTTPPort        string
	UserURL         string
	LedgerAddr      string
	RedisURL        string
	KafkaBrokers    string
	NotificationURL string
}

func Load() Config {
	return Config{
		DatabaseURL:     env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:        env("HTTP_PORT", "8082"),
		UserURL:         env("USER_SERVICE_URL", "http://localhost:8081"),
		LedgerAddr:      env("LEDGER_GRPC_ADDR", "localhost:50051"),
		RedisURL:        env("REDIS_URL", "redis://localhost:6379/0"),
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