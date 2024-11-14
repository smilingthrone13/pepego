package app

import (
	"apubot/internal/config"
	"apubot/internal/handler"
	"apubot/internal/infrastructure/database"
	"apubot/internal/infrastructure/repository"
	"apubot/internal/infrastructure/webapi"
	"apubot/internal/server"
	"apubot/internal/service"
)

type App struct {
	cfg    *config.Config
	server *server.Server
}

func New(cfg *config.Config) *App {
	webAPI := webapi.New(cfg)

	db := database.New(cfg)

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
			APIs:     webAPI,
			Services: services,
		},
	)

	s := server.New(
		&server.InitParams{
			Config:   cfg,
			Api:      webAPI.TgBot,
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
