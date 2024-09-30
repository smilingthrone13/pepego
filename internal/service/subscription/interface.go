package subscription

import (
	"apubot/internal/domain"
	"context"
)

type SubscriptionService interface {
	Get(ctx context.Context, chatId string) (sub domain.Subscription, err error)
	Create(ctx context.Context, sub domain.Subscription) error
	Delete(ctx context.Context, chatId string) error
}

type SubscriptionRepository interface {
	Get(ctx context.Context, chatId string) (sub domain.Subscription, err error)
	GetAll(ctx context.Context) (subs []domain.Subscription, err error)
	Create(ctx context.Context, sub domain.Subscription) error
	Delete(ctx context.Context, chatId string) error
}
