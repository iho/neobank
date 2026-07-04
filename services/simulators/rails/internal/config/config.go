package config

import "os"

type Config struct {
	DatabaseURL   string
	HTTPPort      string
	WebhookURL    string
	WebhookSecret string
	IBANCountry   string
	IBANBankCode  string
}

func Load() Config {
	return Config{
		DatabaseURL:   env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:      env("HTTP_PORT", "8090"),
		WebhookURL:    env("PAYMENT_WEBHOOK_URL", "http://localhost:8082/webhooks/rails"),
		WebhookSecret: env("RAILS_WEBHOOK_SECRET", "dev-rails-webhook-secret"),
		IBANCountry:   env("RAILS_IBAN_COUNTRY", "DE"),
		IBANBankCode:  env("RAILS_IBAN_BANK_CODE", "SIMDEFF"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
