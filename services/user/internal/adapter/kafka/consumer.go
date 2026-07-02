package kafkaadapter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/iho/neobank/pkg/events"
	"github.com/iho/neobank/services/user/internal/usecase"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	brokers string
	groupID string
	project *usecase.ProjectWalletEventUseCase
	logger  *slog.Logger
}

func NewConsumer(brokers, groupID string, project *usecase.ProjectWalletEventUseCase, logger *slog.Logger) *Consumer {
	if groupID == "" {
		groupID = "user-wallet-projection"
	}
	return &Consumer{brokers: brokers, groupID: groupID, project: project, logger: logger}
}

func (c *Consumer) Run(ctx context.Context, topics ...string) error {
	errCh := make(chan error, len(topics))
	for _, topic := range topics {
		topic := topic
		go func() {
			errCh <- c.consumeTopic(ctx, topic)
		}()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func (c *Consumer) consumeTopic(ctx context.Context, topic string) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(c.brokers, ","),
		Topic:   topic,
		GroupID: c.groupID,
	})
	defer reader.Close()

	c.logger.Info("wallet projection consumer started", "topic", topic, "group", c.groupID)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("fetch %s: %w", topic, err)
		}

		var envelope events.Envelope
		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			c.logger.Warn("invalid kafka envelope", "topic", topic, "error", err)
		} else if err := c.project.Execute(ctx, envelope); err != nil {
			c.logger.Warn("wallet projection failed", "topic", topic, "type", envelope.EventType, "error", err)
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit %s: %w", topic, err)
		}
	}
}