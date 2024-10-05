package app

import (
	"apubot/internal/config"
	"apubot/internal/handler"
	"apubot/internal/infrastructure/database"
	"apubot/internal/infrastructure/repository"
	"apubot/internal/server"
	"apubot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type App struct {
	cfg    *config.Config
	server *server.Server
}

func New(cfg *config.Config) *App {
	bot, err := tgbotapi.NewBotAPI(cfg.ApiKey)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.Debug = cfg.IsDebug

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	repos := repository.New(
		&repository.InitParams{
			Config: cfg,
			DB:     db,
		},
	)

	services := service.New(
		&service.InitParams{
			Config:       cfg,
			Repositories: repos,
		},
	)

	handlers := handler.New(
		&handler.InitParams{
			Config:   cfg,
			Bot:      bot,
			Services: services,
		},
	)

	s := server.New(
		&server.InitParams{
			Config:   cfg,
			Bot:      bot,
			Handlers: handlers,
		},
	)

	return &App{
		cfg:    cfg,
		server: s,
	}
}

func (a *App) Run() {
	a.server.Start()
}
