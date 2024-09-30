package repository

import (
	"apubot/internal/config"
	"apubot/internal/infrastructure/database"
	"apubot/internal/infrastructure/repository/image"
	"apubot/internal/infrastructure/repository/subscriprion"
)

type (
	InitParams struct {
		Config *config.Config
		DB     *database.DB
	}

	Repositories struct {
		Image        *image.Repository
		Subscription *subscriprion.Repository
	}
)

func New(p *InitParams) *Repositories {
	return &Repositories{
		Image:        image.New(p.DB),
		Subscription: subscriprion.New(p.DB),
	}
}
