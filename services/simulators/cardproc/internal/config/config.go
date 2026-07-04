package config

import "os"

type Config struct {
	DatabaseURL   string
	HTTPPort      string
	AuthorizeURL  string
	EventsURL     string
	WebhookSecret string
}

func Load() Config {
	return Config{
		DatabaseURL:   env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:      env("HTTP_PORT", "8091"),
		AuthorizeURL:  env("CARD_SERVICE_AUTHORIZE_URL", "http://localhost:8084/webhooks/cardproc/authorize"),
		EventsURL:     env("CARD_SERVICE_EVENTS_URL", "http://localhost:8084/webhooks/cardproc/events"),
		WebhookSecret: env("CARDPROC_WEBHOOK_SECRET", "dev-cardproc-webhook-secret"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
