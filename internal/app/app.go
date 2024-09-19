package app

import (
	"apubot/internal/config"
	"apubot/internal/handler/getter"
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
	lastUsage *cache.Cache
}

func New(cfg *config.Config) *App {
	bot, err := tgbotapi.NewBotAPI(cfg.ApiKey)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	lastUsage := cache.New(cfg.GetterCooldown, 5*time.Minute)

	return &App{
		cfg:       cfg,
		bot:       bot,
		lastUsage: lastUsage,
	}
}

func (a *App) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := a.bot.GetUpdatesChan(u)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case update := <-updates:
			a.handleUpdate(&update)
		case <-c:
			log.Println("Stopping bot...")

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
			msg := tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf("Wait %v seconds before using this command again", a.cfg.GetterCooldown),
			)

			_, err := a.bot.Send(msg)
			if err != nil {
				log.Printf("Error sending message: %v", err)
			}

			return
		}
	}

	switch update.Message.Command() {
	case "peepo":
		go getter.HandleGetCommand(a.bot, update.Message)
	default:
		_, err := a.bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command"))
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}

	a.lastUsage.Set(fmt.Sprint(update.Message.Chat.ID), time.Now(), cache.DefaultExpiration)
}
