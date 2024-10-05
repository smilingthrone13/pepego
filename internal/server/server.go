package server

import (
	"apubot/internal/config"
	"apubot/internal/handler"
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

const (
	StartCommand            = "start"
	PeepoCommand            = "peepo"
	SubscribeCommand        = "sub"
	UnsubscribeCommand      = "unsub"
	SubscriptionInfoCommand = "sub_info"
	HelpCommand             = "help"
)

type Server struct {
	cfg       *config.Config
	bot       *tgbotapi.BotAPI
	handlers  *handler.Handlers
	lastUsage *cache.Cache
	lastCmd   *cache.Cache
}

type InitParams struct {
	Config   *config.Config
	Bot      *tgbotapi.BotAPI
	Handlers *handler.Handlers
}

func New(p *InitParams) *Server {
	return &Server{
		cfg:       p.Config,
		bot:       p.Bot,
		handlers:  p.Handlers,
		lastUsage: cache.New(p.Config.CommandCooldown, 5*time.Minute),
		lastCmd:   cache.New(time.Minute, 5*time.Minute),
	}
}

func (s *Server) Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updatesChan := s.bot.GetUpdatesChan(u)

	for {
		select {
		case update := <-updatesChan:
			go s.handleUpdate(&update)
		case <-c:
			log.Println("Stopping bot...")

			s.bot.StopReceivingUpdates()

			log.Println("Bot gracefully stopped!")

			return
		}
	}
}

func (s *Server) handleUpdate(update *tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if !update.Message.IsCommand() {
		s.handleMessage(update.Message)

		return
	}

	s.handleCommand(update.Message)
}

func (s *Server) handleMessage(message *tgbotapi.Message) {
	var err error

	lastUsedCmd, _ := s.lastCmd.Get(fmt.Sprint(message.Chat.ID))

	switch lastUsedCmd {
	case SubscribeCommand:
		ctx := context.Background()
		err = s.handlers.Image.CreateSubscription(ctx, message)
	default:
		msgText := "I can only handle listed commands in this chat!"
		s.handlers.General.MessageResponse(message.Chat.ID, msgText)
	}

	if err != nil {
		return
	}

	s.lastCmd.Delete(fmt.Sprint(message.Chat.ID))
}

func (s *Server) handleCommand(message *tgbotapi.Message) {
	if lastTime, ok := s.lastUsage.Get(fmt.Sprint(message.Chat.ID)); ok {
		waitTime := s.cfg.CommandCooldown - time.Since(lastTime.(time.Time))
		if waitTime > 0 {
			msgText := fmt.Sprintf("Command on cooldown for %.1f sec", waitTime.Seconds())
			s.handlers.General.MessageResponse(message.Chat.ID, msgText)

			return
		}
	}

	switch message.Command() {
	case StartCommand:
		s.handlers.General.StartResponse(message.Chat.ID)
	case PeepoCommand:
		ctx := context.Background()
		s.handlers.Image.GetImage(ctx, message)
	case SubscribeCommand:
		ctx := context.Background()
		_ = s.handlers.Image.CreateSubscription(ctx, message)
	case UnsubscribeCommand:
		ctx := context.Background()
		s.handlers.Image.DeleteSubscription(ctx, message)
	case SubscriptionInfoCommand:
		ctx := context.Background()
		s.handlers.Image.GetSubscription(ctx, message)
	case HelpCommand:
		s.handlers.General.HelpResponse(message.Chat.ID)
	default:
		s.handlers.General.MessageResponse(message.Chat.ID, "Unknown command")
	}

	s.lastUsage.Set(fmt.Sprint(message.Chat.ID), time.Now(), cache.DefaultExpiration)
	s.lastCmd.Set(fmt.Sprint(message.Chat.ID), message.Command(), cache.DefaultExpiration)
}
