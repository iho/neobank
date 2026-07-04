package config

import "os"

type Config struct {
	DatabaseURL   string
	HTTPPort      string
	EventsURL     string
	WebhookSecret string
}

func Load() Config {
	return Config{
		DatabaseURL:   env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:      env("HTTP_PORT", "8092"),
		EventsURL:     env("USER_SERVICE_EVENTS_URL", "http://localhost:8081/webhooks/kyc/events"),
		WebhookSecret: env("KYC_WEBHOOK_SECRET", "dev-kyc-webhook-secret"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
