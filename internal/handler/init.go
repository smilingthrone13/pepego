package handler

import (
	"apubot/internal/config"
	generalH "apubot/internal/handler/general"
	imageH "apubot/internal/handler/image"
	"apubot/internal/infrastructure/webapi"
	"apubot/internal/service"
)

type (
	InitParams struct {
		Config   *config.Config
		APIs     *webapi.WebAPIs
		Services *service.Services
	}

	Handlers struct {
		General *generalH.Handler
		Image   *imageH.Handler
	}
)

func New(p *InitParams) *Handlers {
	generalHandler := generalH.New(p.Config, p.APIs.TgBot)

	imageHandler := imageH.New(
		p.Config,
		p.APIs.TgBot,
		&imageH.Services{
			Image:        p.Services.Image,
			Subscription: p.Services.Subscription,
		},
	)

	handlers := &Handlers{
		General: generalHandler,
		Image:   imageHandler,
	}

	return handlers
}
