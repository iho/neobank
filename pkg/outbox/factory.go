package outbox

import (
	"context"
	"log/slog"
)

// ProducerConfig selects how outbox events are delivered.
type ProducerConfig struct {
	KafkaBrokers    string
	NotificationURL string
	Logger          *slog.Logger
}

// NewProducer returns Kafka producer when brokers are set, otherwise HTTP or log fallback.
func NewProducer(cfg ProducerConfig) Producer {
	if cfg.KafkaBrokers != "" {
		if cfg.Logger != nil {
			cfg.Logger.Info("outbox using kafka", "brokers", cfg.KafkaBrokers)
		}
		return NewKafkaProducer(cfg.KafkaBrokers)
	}
	if cfg.NotificationURL != "" {
		if cfg.Logger != nil {
			cfg.Logger.Info("outbox using http notification ingest", "url", cfg.NotificationURL)
		}
		return NewHTTPProducer(cfg.NotificationURL)
	}
	return LogProducerFunc(cfg.Logger)
}

type logProducer struct {
	logger *slog.Logger
}

func LogProducerFunc(logger *slog.Logger) Producer {
	return logProducer{logger: logger}
}

func (p logProducer) Produce(_ context.Context, topic, key string, value []byte) error {
	if p.logger != nil {
		p.logger.Info("outbox event", "topic", topic, "key", key, "bytes", len(value))
	}
	return nil
}