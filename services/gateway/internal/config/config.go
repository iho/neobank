package config

import "os"

type Config struct {
	HTTPPort        string
	UserURL         string
	PaymentURL      string
	CardURL         string
	NotificationURL string
	LedgerAddr      string
	RedisURL        string
	JWTSecret       string
}

func Load() Config {
	return Config{
		HTTPPort:        env("HTTP_PORT", "8080"),
		UserURL:         env("USER_SERVICE_URL", "http://localhost:8081"),
		PaymentURL:      env("PAYMENT_SERVICE_URL", "http://localhost:8082"),
		CardURL:         env("CARD_SERVICE_URL", "http://localhost:8084"),
		NotificationURL: env("NOTIFICATION_SERVICE_URL", "http://localhost:8083"),
		LedgerAddr:      env("LEDGER_GRPC_ADDR", "localhost:50051"),
		RedisURL:        env("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:       env("JWT_SECRET", "dev-secret-change-me"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}