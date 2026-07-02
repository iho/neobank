package usecase

import (
	"context"

	"github.com/iho/neobank/pkg/notify"
	"github.com/iho/neobank/pkg/userclient"
)

type DeviceTokenProvider interface {
	ListTokens(ctx context.Context, userID string) ([]string, error)
}

type UserEmailProvider interface {
	Email(ctx context.Context, userID string) (string, error)
}

type UserClientDelivery struct {
	users *userclient.Client
}

func NewUserClientDelivery(users *userclient.Client) *UserClientDelivery {
	return &UserClientDelivery{users: users}
}

func (p *UserClientDelivery) ListTokens(ctx context.Context, userID string) ([]string, error) {
	tokens, err := p.users.ListDeviceTokens(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, t.Token)
	}
	return out, nil
}

func (p *UserClientDelivery) Email(ctx context.Context, userID string) (string, error) {
	user, err := p.users.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.Email, nil
}

type DeliveryService struct {
	dispatcher notify.Dispatcher
	tokens     DeviceTokenProvider
	emails     UserEmailProvider
	prefs      PreferencesRepository
}

func NewDeliveryService(
	dispatcher notify.Dispatcher,
	tokens DeviceTokenProvider,
	emails UserEmailProvider,
	prefs PreferencesRepository,
) *DeliveryService {
	return &DeliveryService{
		dispatcher: dispatcher,
		tokens:     tokens,
		emails:     emails,
		prefs:      prefs,
	}
}

func (s *DeliveryService) Deliver(ctx context.Context, userID, eventType, title, body string) error {
	if s == nil || s.dispatcher == nil {
		return nil
	}

	prefs, err := s.prefs.Get(ctx, userID)
	if err != nil {
		return err
	}
	category := CategoryForEventType(eventType)
	if category != "" && !AllowsCategory(prefs, category) {
		return nil
	}

	if prefs.Push && s.tokens != nil {
		tokens, err := s.tokens.ListTokens(ctx, userID)
		if err == nil && len(tokens) > 0 {
			_ = s.dispatcher.Dispatch(ctx, notify.Message{
				UserID:  userID,
				Title:   title,
				Body:    body,
				Tokens:  tokens,
				Channel: notify.ChannelPush,
			})
		}
	}

	if prefs.Email && s.emails != nil {
		email, err := s.emails.Email(ctx, userID)
		if err == nil && email != "" {
			_ = s.dispatcher.Dispatch(ctx, notify.Message{
				UserID:  userID,
				Title:   title,
				Body:    body,
				Email:   email,
				Channel: notify.ChannelEmail,
			})
		}
	}

	return nil
}