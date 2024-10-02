package general

import (
	"apubot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type (
	Handler struct {
		cfg *config.Config
		bot *tgbotapi.BotAPI
	}
)

func New(cfg *config.Config, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		cfg: cfg,
		bot: bot,
	}
}

func (h *Handler) MessageResponse(chatID int64, message string) {
	_, err := h.bot.Send(tgbotapi.NewMessage(chatID, message))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) StartResponse(chatID int64) {
	msgText := "Welcome to peepobot. Now you can use any available command."

	_, err := h.bot.Send(tgbotapi.NewMessage(chatID, msgText))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) HelpResponse(chatID int64) {
	msgText := "Command list help:\n" +
		"/peepo - Get random peepo picture;\n" +
		"/sub <period> - Subscribe to receive peepo pictures. Example: \"/sub 1h30m20s\";\n" +
		"/sub_info - Get info about current subscription;\n" +
		"/unsub - Unsubscribe from receiving peepo pictures;\n" +
		"/help - Get this list."

	_, err := h.bot.Send(tgbotapi.NewMessage(chatID, msgText))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
