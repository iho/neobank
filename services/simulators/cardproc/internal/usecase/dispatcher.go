package usecase

import "context"

// WebhookDispatcher schedules a signed webhook delivery; satisfied by
// pkg/vendorsim.Dispatcher.
type WebhookDispatcher interface {
	Enqueue(ctx context.Context, url, eventType string, payload any) (string, error)
}
