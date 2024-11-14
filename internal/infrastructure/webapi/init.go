package webapi

import (
	"apubot/internal/config"
	"apubot/internal/infrastructure/webapi/tg_bot"
)

type WebAPIs struct {
	TgBot *tg_bot.BotAPI
}

func New(cfg *config.Config) *WebAPIs {
	tgBot := tg_bot.New(cfg)

	return &WebAPIs{
		TgBot: tgBot,
	}
}
