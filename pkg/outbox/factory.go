//
// Copyright (c) 2026 Sumicare
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package outbox

import (
	"context"
	"log/slog"
)

// ProducerConfig selects how outbox events are delivered.
type ProducerConfig struct {
	Logger          *slog.Logger
	KafkaBrokers    string
	NotificationURL string
	ProjectionURLs  []string
}

// NewProducer returns Kafka producer when brokers are set, otherwise HTTP or log fallback.
func NewProducer(cfg ProducerConfig) Producer {
	if cfg.KafkaBrokers != "" {
		if cfg.Logger != nil {
			cfg.Logger.Info("outbox using kafka", "brokers", cfg.KafkaBrokers)
		}

		return NewKafkaProducer(cfg.KafkaBrokers)
	}

	var producers []Producer
	if cfg.NotificationURL != "" {
		if cfg.Logger != nil {
			cfg.Logger.Info("outbox using http notification ingest", "url", cfg.NotificationURL)
		}

		producers = append(producers, NewHTTPProducer(cfg.NotificationURL))
	}

	for _, url := range cfg.ProjectionURLs {
		if url == "" {
			continue
		}

		if cfg.Logger != nil {
			cfg.Logger.Info("outbox using http projection ingest", "url", url)
		}

		producers = append(producers, NewHTTPProducer(url))
	}

	if len(producers) == 1 {
		return producers[0]
	}

	if len(producers) > 1 {
		return NewMultiProducer(producers...)
	}

	return LogProducerFunc(cfg.Logger)
}

type logProducer struct {
	logger *slog.Logger
}

func LogProducerFunc(logger *slog.Logger) logProducer {
	return logProducer{logger: logger}
}

func (p logProducer) Produce(_ context.Context, topic, key string, value []byte) error {
	if p.logger != nil {
		p.logger.Info("outbox event", "topic", topic, "key", key, "bytes", len(value))
	}

	return nil
}
