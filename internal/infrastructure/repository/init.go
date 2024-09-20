package repository

import (
	"apubot/internal/config"
	"apubot/internal/infrastructure/database"
	"apubot/internal/infrastructure/repository/image"
)

type (
	InitParams struct {
		Config *config.Config
		DB     *database.DB
	}

	Repositories struct {
		Image *image.Repository
	}
)

func New(p *InitParams) *Repositories {
	return &Repositories{
		Image: image.New(p.DB),
	}
}
