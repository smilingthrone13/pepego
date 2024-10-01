package image

import (
	"apubot/internal/config"
	"apubot/internal/domain"
	"apubot/internal/service/image"
	"apubot/internal/service/subscription"
	"apubot/pkg/custom_errors"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

type (
	Handler struct {
		cfg      *config.Config
		bot      *tgbotapi.BotAPI
		services *Services
	}
	Services struct {
		Image        image.ImageService
		Subscription subscription.SubscriptionService
	}
)

func New(cfg *config.Config, bot *tgbotapi.BotAPI, services *Services) *Handler {
	h := &Handler{
		cfg:      cfg,
		bot:      bot,
		services: services,
	}

	err := h.services.Subscription.RescheduleExisting(context.Background(), h.sendImage)
	if err != nil {
		log.Fatal(err)
	}

	return h
}

func (h *Handler) GetImage(ctx context.Context, message *tgbotapi.Message) {
	file, err := h.services.Image.GetRandomFile(ctx)
	if err != nil {
		log.Printf("Error getting file: %v", err)

		return
	}

	attachment, err := h.createAttachment(file, message.Chat.ID)
	if err != nil {
		log.Printf("Error creating attachment: %v", err)

		return
	}

	res, err := h.bot.Send(attachment)
	if err != nil {
		log.Printf("Error sending attachment: %v", err)

		msg := tgbotapi.NewMessage(message.Chat.ID, "Can not send an image monkaS")
		_, err = h.bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return
	}

	if file.TgID == "" {
		h.updateFile(ctx, file, res)
	}
}

func (h *Handler) CreateSubscription(ctx context.Context, message *tgbotapi.Message) {
	inp, err := parseAndValidateSubscriptionInput(message)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, err.Error())
		_, err = h.bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return
	}

	err = h.services.Subscription.Create(ctx, inp, h.sendImage)
	if err != nil {
		log.Printf("Error creating subscription: %v", err)

		return
	}

	msgText := "Subscription created successfully"
	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) DeleteSubscription(ctx context.Context, message *tgbotapi.Message) {
	sub, err := h.services.Subscription.Get(ctx, message.Chat.ID)
	if err != nil {
		msgText := "Error getting subscription :d"

		var notFoundErr *custom_errors.NotFoundError
		if errors.As(err, &notFoundErr) {
			msgText = "No active subscription found!"
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		_, err = h.bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return
	}

	err = h.services.Subscription.Delete(ctx, sub.ChatId)
	if err != nil {
		msgText := "Can not delete subscription :d"
		msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
		_, err = h.bot.Send(msg)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}

		return
	}

	msgText := "Subscription deleted successfully!"
	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
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

func (h *Handler) updateFile(ctx context.Context, file domain.File, res tgbotapi.Message) {
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

	err := h.services.Image.UpdateFile(ctx, updInp)
	if err != nil {
		log.Printf("Error updating file: %v", err)
	}
}

// sendImage is used as an injected function to subscription service
func (h *Handler) sendImage(chatId int64) error {
	ctx := context.Background()

	file, err := h.services.Image.GetRandomFile(ctx)
	if err != nil {
		return err
	}

	attachment, err := h.createAttachment(file, chatId)
	if err != nil {
		return err
	}

	res, err := h.bot.Send(attachment)
	if err != nil {
		return err
	}

	if file.TgID == "" {
		h.updateFile(ctx, file, res)
	}

	return nil
}

func parseAndValidateSubscriptionInput(message *tgbotapi.Message) (domain.Subscription, error) {
	args := message.CommandArguments()
	period, err := strconv.Atoi(args)
	if err != nil {
		err = errors.New("Bad subscription period! Please enter a number between 1 and 24.")

		return domain.Subscription{}, err
	}

	if period <= 0 || period > 24 {
		err = errors.New("Subscription period must be between 1 and 24.")

		return domain.Subscription{}, err
	}

	inp := domain.Subscription{
		ChatId:    message.Chat.ID,
		CreatedAt: time.Now().Unix(),
		Period:    period,
	}

	return inp, nil
}
