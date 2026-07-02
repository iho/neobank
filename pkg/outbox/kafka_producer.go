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
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaProducer publishes outbox envelopes to Kafka topics.
type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(brokers string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(strings.Split(brokers, ",")...),
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
	}
}

func (p *KafkaProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
		Time:  time.Now().UTC(),
	}
	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("kafka produce: %w", err)
	}

	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}
