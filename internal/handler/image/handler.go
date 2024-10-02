package image

import (
	"apubot/internal/config"
	"apubot/internal/domain"
	"apubot/internal/service/image"
	"apubot/internal/service/subscription"
	"apubot/pkg/custom_errors"
	"apubot/pkg/utils/queue"
	"apubot/pkg/utils/time_string"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"path"
	"path/filepath"
	"strings"
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

		return
	}

	if file.TgID == "" {
		h.updateFile(ctx, file, res)
	}
}

func (h *Handler) CreateSubscription(ctx context.Context, message *tgbotapi.Message) {
	inp, err := h.parseAndValidateSubscriptionInput(message)
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

	msgText := "Subscription created successfully!"
	msg := tgbotapi.NewMessage(message.Chat.ID, msgText)
	_, err = h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

func (h *Handler) GetSubscription(ctx context.Context, message *tgbotapi.Message) {
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

	createdAt := sub.SubscribedAtAsUnixTime().String()
	period := time_string.ShortDur(sub.PeriodAsDurationInSeconds())
	passedIntervals := time.Since(sub.SubscribedAtAsUnixTime()) / sub.PeriodAsDurationInSeconds()
	nextEvent := sub.SubscribedAtAsUnixTime().Add((passedIntervals + 1) * sub.PeriodAsDurationInSeconds())

	msgText := "Current subscription info:\n" +
		fmt.Sprintf("Created at: %s\n", createdAt) +
		fmt.Sprintf("Period: %s\n", period) +
		fmt.Sprintf("Next peepo: %s", nextEvent)

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
func (h *Handler) sendImage(chatId int64, q *queue.Queue) error {
	ctx := context.Background()

	var file domain.File
	var err error

	for {
		file, err = h.services.Image.GetRandomFile(ctx)
		if err != nil {
			return err
		}

		// check if file is already in sent queue
		if !q.Contains(file.Name) {
			break
		}
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

	q.Add(file.Name)

	return nil
}

func (h *Handler) parseAndValidateSubscriptionInput(message *tgbotapi.Message) (domain.Subscription, error) {
	args := strings.ReplaceAll(message.CommandArguments(), " ", "")
	period, err := time.ParseDuration(args)
	if err != nil {
		errText := fmt.Sprintf(
			"Please enter a period in format like \"1h30m25s\"! (min: %s, max: %s)",
			time_string.ShortDur(h.cfg.MinSubscriptionPeriod),
			time_string.ShortDur(h.cfg.MaxSubscriptionPeriod),
		)
		err = errors.New(errText)

		return domain.Subscription{}, err
	}

	if period.Seconds() < h.cfg.MinSubscriptionPeriod.Seconds() ||
		period.Seconds() > h.cfg.MaxSubscriptionPeriod.Seconds() {
		errText := fmt.Sprintf(
			"Subscription period must be between %s and %s!",
			time_string.ShortDur(h.cfg.MinSubscriptionPeriod),
			time_string.ShortDur(h.cfg.MaxSubscriptionPeriod),
		)
		err = errors.New(errText)

		return domain.Subscription{}, err
	}

	inp := domain.Subscription{
		ChatId:    message.Chat.ID,
		CreatedAt: time.Now().Unix(),
		Period:    int(period.Seconds()),
	}

	return inp, nil
}
