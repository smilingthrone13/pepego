package server

import (
	"apubot/internal/config"
	"apubot/internal/handler"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/patrickmn/go-cache"
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

type botApi interface {
	GetUpdatesChan() tgbotapi.UpdatesChannel
	Shutdown()
}

type (
	InitParams struct {
		Config   *config.Config
		Api      botApi
		Handlers *handler.Handlers
	}

	Server struct {
		cfg       *config.Config
		api       botApi
		handlers  *handler.Handlers
		lastUsage *cache.Cache
		lastCmd   *cache.Cache
	}
)

func New(p *InitParams) *Server {
	return &Server{
		cfg:       p.Config,
		api:       p.Api,
		handlers:  p.Handlers,
		lastUsage: cache.New(p.Config.CommandCooldown, 5*time.Minute),
		lastCmd:   cache.New(time.Minute, 5*time.Minute),
	}
}

func (s *Server) Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	updatesChan := s.api.GetUpdatesChan()

	for {
		select {
		case update := <-updatesChan:
			go s.handleUpdate(&update)
		case <-c:
			s.api.Shutdown()

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
		err = s.handlers.Image.CreateSubscription(context.Background(), message)
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
		s.handlers.Image.GetImage(context.Background(), message)
	case SubscribeCommand:
		_ = s.handlers.Image.CreateSubscription(context.Background(), message)
	case UnsubscribeCommand:
		s.handlers.Image.DeleteSubscription(context.Background(), message)
	case SubscriptionInfoCommand:
		s.handlers.Image.GetSubscription(context.Background(), message)
	case HelpCommand:
		s.handlers.General.HelpResponse(message.Chat.ID)
	default:
		s.handlers.General.MessageResponse(message.Chat.ID, "Unknown command")
	}

	s.lastUsage.Set(fmt.Sprint(message.Chat.ID), time.Now(), cache.DefaultExpiration)
	s.lastCmd.Set(fmt.Sprint(message.Chat.ID), message.Command(), cache.DefaultExpiration)
}
