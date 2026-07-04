package config

import "os"

type Config struct {
	DatabaseURL string
	HTTPPort    string
}

func Load() Config {
	return Config{
		DatabaseURL: env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:    env("HTTP_PORT", "8093"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
