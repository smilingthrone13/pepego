package handler

import (
	"apubot/internal/config"
	getterH "apubot/internal/handler/getter"
	"apubot/internal/service"
)

type (
	InitParams struct {
		Config   *config.Config
		Services *service.Services
	}

	Handlers struct {
		Getter *getterH.Handler
	}
)

func New(p *InitParams) *Handlers {
	return &Handlers{
		Getter: getterH.New(p.Config, &getterH.Services{Image: p.Services.Image}),
	}
}
