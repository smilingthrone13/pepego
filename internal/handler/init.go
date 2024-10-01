package handler

import (
	"apubot/internal/config"
	getterG "apubot/internal/handler/general"
	getterI "apubot/internal/handler/image"
	"apubot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type (
	InitParams struct {
		Config   *config.Config
		Bot      *tgbotapi.BotAPI
		Services *service.Services
	}

	Handlers struct {
		General *getterG.Handler
		Image   *getterI.Handler
	}
)

func New(p *InitParams) *Handlers {
	return &Handlers{
		General: getterG.New(p.Config, p.Bot),
		Image: getterI.New(
			p.Config,
			p.Bot,
			&getterI.Services{
				Image:        p.Services.Image,
				Subscription: p.Services.Subscription,
			},
		),
	}
}
