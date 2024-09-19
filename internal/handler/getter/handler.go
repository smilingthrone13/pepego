package getter

import (
	"apubot/internal/config"
	"apubot/internal/service/image"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"path"
	"path/filepath"
)

type (
	Handler struct {
		cfg      *config.Config
		Services Services
	}
	Services struct {
		Image image.ImageService
	}
)

func New(cfg *config.Config, services Services) *Handler {
	return &Handler{
		cfg:      cfg,
		Services: services,
	}
}

func (h *Handler) HandleGetCommand(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	attachment, err := h.createAttachment(ctx, message.Chat.ID)
	if err != nil {
		log.Printf("Error creating attachment: %v", err)
	}

	_, err = bot.Send(attachment)
	if err != nil {
		log.Printf("Error sending attachment: %v", err)

		msg := tgbotapi.NewMessage(message.Chat.ID, "Can not send an image monkaS")

		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

func (h *Handler) createAttachment(ctx context.Context, chatId int64) (tgbotapi.Chattable, error) {
	var attachment tgbotapi.Chattable
	var err error

	file := h.Services.Image.GetRandomFile(ctx)
	fullFilePath := path.Join(h.cfg.ImagesDirPath, file.Name)

	switch filepath.Ext(file.Name) {
	case ".jpg", ".jpeg", ".png":
		attachment = tgbotapi.NewPhoto(
			chatId,
			tgbotapi.FilePath(fullFilePath),
		)
	case ".gif":
		attachment = tgbotapi.NewDocument(
			chatId,
			tgbotapi.FilePath(fullFilePath),
		)
	default:
		err = fmt.Errorf("unsupported image format: %v", filepath.Ext(file.Name))
	}

	return attachment, err
}
