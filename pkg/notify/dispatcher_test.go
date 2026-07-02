package notify_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/iho/neobank/pkg/notify"
)

func TestLogDispatcherDispatch(t *testing.T) {
	d := notify.NewLogDispatcher(slog.Default())
	if err := d.Dispatch(context.Background(), notify.Message{
		UserID:  "user-1",
		Title:   "Hello",
		Body:    "World",
		Email:   "a@example.com",
		Tokens:  []string{"tok-1"},
		Channel: notify.ChannelPush,
	}); err != nil {
		t.Fatalf("dispatch push: %v", err)
	}
	if err := d.Dispatch(context.Background(), notify.Message{
		UserID:  "user-1",
		Title:   "Hello",
		Body:    "World",
		Email:   "a@example.com",
		Channel: notify.ChannelEmail,
	}); err != nil {
		t.Fatalf("dispatch email: %v", err)
	}
}