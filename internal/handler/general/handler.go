package general

import (
	"apubot/internal/config"
)

type (
	botApi interface {
		SendMessage(chatID int64, message string)
	}

	Handler struct {
		cfg *config.Config
		api botApi
	}
)

func New(cfg *config.Config, botAPI botApi) *Handler {
	return &Handler{
		cfg: cfg,
		api: botAPI,
	}
}

func (h *Handler) MessageResponse(chatID int64, message string) {
	h.api.SendMessage(chatID, message)
}

func (h *Handler) StartResponse(chatID int64) {
	message := "Welcome to peepobot. Now you can use any available command."

	h.api.SendMessage(chatID, message)
}

func (h *Handler) HelpResponse(chatID int64) {
	message := "Command list help:\n" +
		"/peepo - Get random picture;\n" +
		"/sub - Subscribe to receive pictures periodically;\n" +
		"/sub_info - Get info about current subscription;\n" +
		"/unsub - Drop current subscription;\n" +
		"/help - Get this list."

	h.api.SendMessage(chatID, message)
}
