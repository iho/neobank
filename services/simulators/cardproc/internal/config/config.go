package config

import (
	"os"
	"time"
)

type Config struct {
	DatabaseURL   string
	HTTPPort      string
	AuthorizeURL  string
	EventsURL     string
	WebhookSecret string
	// AuthTTL is how long an approved-but-uncaptured hold sits before the
	// background sweep expires it. Real processors hold auths for days;
	// this defaults short so the expiry flow is actually observable in a
	// demo/test run rather than requiring a multi-day wait.
	AuthTTL time.Duration
	// AuthSweepInterval is how often the expiry sweep runs.
	AuthSweepInterval time.Duration
}

func Load() Config {
	return Config{
		DatabaseURL:       env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:          env("HTTP_PORT", "8091"),
		AuthorizeURL:      env("CARD_SERVICE_AUTHORIZE_URL", "http://localhost:8084/webhooks/cardproc/authorize"),
		EventsURL:         env("CARD_SERVICE_EVENTS_URL", "http://localhost:8084/webhooks/cardproc/events"),
		WebhookSecret:     env("CARDPROC_WEBHOOK_SECRET", "dev-cardproc-webhook-secret"),
		AuthTTL:           envDuration("CARDPROC_AUTH_TTL", 5*time.Minute),
		AuthSweepInterval: envDuration("CARDPROC_AUTH_SWEEP_INTERVAL", 30*time.Second),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}

	return fallback
}
