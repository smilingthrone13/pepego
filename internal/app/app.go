package app

import (
	"apubot/internal/config"
	"apubot/internal/handler"
	"apubot/internal/infrastructure/database"
	"apubot/internal/infrastructure/repository"
	"apubot/internal/service"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/patrickmn/go-cache"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	cfg       *config.Config
	bot       *tgbotapi.BotAPI
	db        *database.DB
	handlers  *handler.Handlers
	lastUsage *cache.Cache
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

	lastUsage := cache.New(cfg.GetterCooldown, 5*time.Minute)

	return &App{
		cfg:       cfg,
		bot:       bot,
		db:        db,
		handlers:  handlers,
		lastUsage: lastUsage,
	}
}

func (a *App) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updatesChan := a.bot.GetUpdatesChan(u)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case update := <-updatesChan:
			a.handleUpdate(&update)
		case <-c:
			log.Println("Stopping bot...")

			a.bot.StopReceivingUpdates()

			err := a.db.Close()
			if err != nil {
				log.Printf("Error closing database conn: %v\n", err)
			}

			log.Println("Bot gracefully stopped!")

			return
		}
	}
}

func (a *App) handleUpdate(update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if !update.Message.IsCommand() {
		return
	}

	if lastTime, ok := a.lastUsage.Get(fmt.Sprint(update.Message.Chat.ID)); ok {
		waitTime := a.cfg.GetterCooldown - time.Since(lastTime.(time.Time))
		if waitTime > 0 {
			msgText := fmt.Sprintf("Command on cooldown for %.1f sec", waitTime.Seconds())

			go a.handlers.General.MessageResponse(update.Message.Chat.ID, msgText)

			return
		}
	}

	switch update.Message.Command() {
	case "start":
		msgText := "Welcome to peepobot. Now you can use any available command."
		go a.handlers.General.MessageResponse(update.Message.Chat.ID, msgText)
	case "peepo":
		ctx := context.Background()
		go a.handlers.Image.GetImage(ctx, update.Message)
	case "subscribe":
		ctx := context.Background()
		go a.handlers.Image.CreateSubscription(ctx, update.Message)
	case "unsubscribe":
		ctx := context.Background()
		go a.handlers.Image.DeleteSubscription(ctx, update.Message)
	case "help":
		go a.handlers.General.HelpResponse(update.Message.Chat.ID)
	default:
		go a.handlers.General.MessageResponse(update.Message.Chat.ID, "Unknown command")
	}

	a.lastUsage.Set(fmt.Sprint(update.Message.Chat.ID), time.Now(), cache.DefaultExpiration)
}
