package tg_bot

import (
	"apubot/internal/config"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type BotAPI struct {
	bot *tgbotapi.BotAPI
}

func New(cfg *config.Config) *BotAPI {
	bot, err := tgbotapi.NewBotAPI(cfg.ApiKey)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	bot.Debug = cfg.IsDebug

	return &BotAPI{
		bot: bot,
	}
}

func (b *BotAPI) SendMessage(chatID int64, message string) {
	_, err := b.bot.Send(tgbotapi.NewMessage(chatID, message))
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (b *BotAPI) SendAttachment(attachment tgbotapi.Chattable) (res tgbotapi.Message, err error) {
	res, err = b.bot.Send(attachment)
	if err != nil {
		log.Printf("Error sending attachment: %v", err)

		return tgbotapi.Message{}, err
	}

	return res, nil
}

func (b *BotAPI) GetUpdatesChan() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.bot.GetUpdatesChan(u)
}

func (b *BotAPI) Shutdown() {
	log.Println("Stopping bot...")

	b.bot.StopReceivingUpdates()
}
