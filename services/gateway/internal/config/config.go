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
	AppEnv          string
	// AllowDevAuth gates the X-User-Id header bypass and the legacy
	// access.<user-id>.<anything> token: both skip real JWT validation and
	// must never be reachable outside local development.
	AllowDevAuth bool
}

func Load() Config {
	appEnv := env("APP_ENV", "development")
	return Config{
		HTTPPort:        env("HTTP_PORT", "8080"),
		UserURL:         env("USER_SERVICE_URL", "http://localhost:8081"),
		PaymentURL:      env("PAYMENT_SERVICE_URL", "http://localhost:8082"),
		CardURL:         env("CARD_SERVICE_URL", "http://localhost:8084"),
		NotificationURL: env("NOTIFICATION_SERVICE_URL", "http://localhost:8083"),
		LedgerAddr:      env("LEDGER_GRPC_ADDR", "localhost:50051"),
		RedisURL:        env("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:       env("JWT_SECRET", "dev-secret-change-me"),
		AppEnv:          appEnv,
		AllowDevAuth:    appEnv == "development" || appEnv == "local" || appEnv == "dev",
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
