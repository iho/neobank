package config

import "os"

type Config struct {
	DatabaseURL  string
	HTTPPort     string
	KafkaBrokers string
	UserURL      string
}

func Load() Config {
	return Config{
		DatabaseURL:  env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:     env("HTTP_PORT", "8083"),
		KafkaBrokers: env("KAFKA_BROKERS", ""),
		UserURL:      env("USER_URL", "http://localhost:8081"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}