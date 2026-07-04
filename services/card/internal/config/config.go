package config

import "os"

type Config struct {
	DatabaseURL            string
	HTTPPort               string
	GRPCPort               string
	UserURL                string
	UserGRPCAddr           string
	LedgerAddr             string
	RedisURL               string
	KafkaBrokers           string
	SettlementLedgerAcctID string
	NotificationURL        string
	CardProcURL            string
	CardProcWebhookSecret  string
}

func Load() Config {
	return Config{
		DatabaseURL:            env("DATABASE_URL", "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"),
		HTTPPort:               env("HTTP_PORT", "8084"),
		GRPCPort:               env("GRPC_PORT", "50054"),
		UserURL:                env("USER_SERVICE_URL", "http://localhost:8081"),
		UserGRPCAddr:           env("USER_GRPC_ADDR", "localhost:50052"),
		LedgerAddr:             env("LEDGER_GRPC_ADDR", "localhost:50051"),
		RedisURL:               env("REDIS_URL", "redis://localhost:6379/0"),
		KafkaBrokers:           env("KAFKA_BROKERS", ""),
		SettlementLedgerAcctID: env("SETTLEMENT_LEDGER_ACCOUNT_ID", ""),
		NotificationURL:        env("NOTIFICATION_SERVICE_URL", "http://localhost:8083/api/v1/internal/events"),
		CardProcURL:            env("CARDPROC_SERVICE_URL", "http://localhost:8091"),
		CardProcWebhookSecret:  env("CARDPROC_WEBHOOK_SECRET", "dev-cardproc-webhook-secret"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
