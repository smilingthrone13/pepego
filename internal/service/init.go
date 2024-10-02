package service

import (
	"apubot/internal/config"
	"apubot/internal/infrastructure/repository"
	"apubot/internal/service/image"
	"apubot/internal/service/subscription"
)

type (
	InitParams struct {
		Config       *config.Config
		Repositories *repository.Repositories
	}

	Services struct {
		Image        *image.Service
		Subscription *subscription.Service
	}
)

func New(p *InitParams) *Services {
	return &Services{
		Image:        image.New(p.Config, p.Repositories.Image),
		Subscription: subscription.New(p.Config, p.Repositories.Subscription),
	}
}
