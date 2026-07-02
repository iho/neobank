package config

import "os"

type Config struct {
	HTTPPort             string
	UserGRPCAddr         string
	PaymentGRPCAddr      string
	CardGRPCAddr         string
	NotificationGRPCAddr string
	LedgerAddr           string
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
		HTTPPort:             env("HTTP_PORT", "8080"),
		UserGRPCAddr:         env("USER_GRPC_ADDR", "localhost:50052"),
		PaymentGRPCAddr:      env("PAYMENT_GRPC_ADDR", "localhost:50053"),
		CardGRPCAddr:         env("CARD_GRPC_ADDR", "localhost:50054"),
		NotificationGRPCAddr: env("NOTIFICATION_GRPC_ADDR", "localhost:50055"),
		LedgerAddr:           env("LEDGER_GRPC_ADDR", "localhost:50051"),
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
