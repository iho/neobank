package config

import "os"

type Config struct {
	DatabaseURL  string
	HTTPPort     string
	GRPCPort     string
	KafkaBrokers string
	UserGRPCAddr string
}

func Load() Config {
	return Config{
		DatabaseURL:  env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:     env("HTTP_PORT", "8083"),
		GRPCPort:     env("GRPC_PORT", "50055"),
		KafkaBrokers: env("KAFKA_BROKERS", ""),
		UserGRPCAddr: env("USER_GRPC_ADDR", "localhost:50052"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}