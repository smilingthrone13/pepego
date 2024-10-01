package subscription

import (
	"apubot/internal/domain"
	"context"
)

type SubscriptionService interface {
	Get(ctx context.Context, chatId int64) (sub domain.Subscription, err error)
	Create(ctx context.Context, sub domain.Subscription, sendFunc func(chatId int64) error) error
	Delete(ctx context.Context, chatId int64) error
	RescheduleExisting(ctx context.Context, sendFunc func(chatId int64) error) error
}

type SubscriptionRepository interface {
	Get(ctx context.Context, chatId int64) (sub domain.Subscription, err error)
	GetAll(ctx context.Context) (subs []domain.Subscription, err error)
	Create(ctx context.Context, sub domain.Subscription) error
	Delete(ctx context.Context, chatId int64) error
}
