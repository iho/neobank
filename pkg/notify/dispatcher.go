package notify

import (
	"context"
	"log/slog"
)

// Channel identifies an outbound delivery channel.
type Channel string

const (
	ChannelPush  Channel = "push"
	ChannelEmail Channel = "email"
)

// Message is a notification ready for external delivery.
type Message struct {
	UserID  string
	Title   string
	Body    string
	Email   string
	Tokens  []string
	Channel Channel
}

// Dispatcher delivers notifications outside the in-app inbox.
type Dispatcher interface {
	Dispatch(ctx context.Context, msg Message) error
}

// LogDispatcher logs push/email attempts (MVP stub for FCM/SMTP).
type LogDispatcher struct {
	logger *slog.Logger
}

func NewLogDispatcher(logger *slog.Logger) *LogDispatcher {
	if logger == nil {
		logger = slog.Default()
	}
	return &LogDispatcher{logger: logger}
}

func (d *LogDispatcher) Dispatch(ctx context.Context, msg Message) error {
	switch msg.Channel {
	case ChannelPush:
		d.logger.InfoContext(ctx, "push notification dispatched",
			"user_id", msg.UserID,
			"title", msg.Title,
			"token_count", len(msg.Tokens),
		)
	case ChannelEmail:
		d.logger.InfoContext(ctx, "email notification dispatched",
			"user_id", msg.UserID,
			"email", msg.Email,
			"title", msg.Title,
		)
	default:
		d.logger.InfoContext(ctx, "notification dispatched",
			"user_id", msg.UserID,
			"channel", msg.Channel,
			"title", msg.Title,
		)
	}
	return nil
}

// MultiDispatcher fans out to every configured channel.
type MultiDispatcher struct {
	dispatchers []Dispatcher
}

func NewMultiDispatcher(dispatchers ...Dispatcher) *MultiDispatcher {
	return &MultiDispatcher{dispatchers: dispatchers}
}

func (m *MultiDispatcher) Dispatch(ctx context.Context, msg Message) error {
	for _, d := range m.dispatchers {
		if err := d.Dispatch(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}