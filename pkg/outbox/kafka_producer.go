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
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka produce: %w", err)
	}
	return nil
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}