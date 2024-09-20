package getter

import (
	"apubot/internal/config"
	"apubot/internal/domain"
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
		Services *Services
	}
	Services struct {
		Image image.ImageService
	}
)

func New(cfg *config.Config, services *Services) *Handler {
	return &Handler{
		cfg:      cfg,
		Services: services,
	}
}

func (h *Handler) HandleGetCommand(ctx context.Context, bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	file := h.Services.Image.GetRandomFile(ctx)

	attachment, err := h.createAttachment(file, message.Chat.ID)
	if err != nil {
		log.Printf("Error creating attachment: %v", err)

		return
	}

	res, err := bot.Send(attachment)
	if err != nil {
		log.Printf("Error sending attachment: %v", err)

		msg := tgbotapi.NewMessage(message.Chat.ID, "Can not send an image monkaS")

		_, err = bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return
	}

	_ = res

	if file.TgID != "" {
		return
	}

	var newTgId string

	switch filepath.Ext(file.Name) {
	case ".jpg", ".jpeg", ".png":
		if res.Photo == nil || len(res.Photo) == 0 {
			log.Println("Photo is nil in response!")

			return
		}

		maxSizedImage := res.Photo[len(res.Photo)-1]
		newTgId = maxSizedImage.FileID
	case ".gif":
		if res.Animation == nil {
			log.Println("Animation is nil in response!")

			return
		}

		newTgId = res.Animation.FileID
	default:
		log.Printf("Unsupported image format: %v", filepath.Ext(file.Name))
	}

	if newTgId == "" {
		log.Println("No new TG ID in response!")

		return
	}

	updInp := domain.File{Name: file.Name, TgID: newTgId}

	err = h.Services.Image.UpdateFile(ctx, updInp)
	if err != nil {
		log.Printf("Error updating file: %v", err)
	}
}

func (h *Handler) createAttachment(file domain.File, chatId int64) (a tgbotapi.Chattable, err error) {
	var reqFile tgbotapi.RequestFileData

	if file.TgID == "" {
		fullFilePath := path.Join(h.cfg.ImagesDirPath, file.Name)
		reqFile = tgbotapi.FilePath(fullFilePath)
	} else {
		reqFile = tgbotapi.FileID(file.TgID)
	}

	switch filepath.Ext(file.Name) {
	case ".jpg", ".jpeg", ".png":
		a = tgbotapi.NewPhoto(chatId, reqFile)
	case ".gif":
		a = tgbotapi.NewDocument(chatId, reqFile)
	default:
		err = fmt.Errorf("unsupported image format: %v", filepath.Ext(file.Name))
	}

	return a, err
}
