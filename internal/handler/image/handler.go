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

type botApi interface {
	SendMessage(chatID int64, message string)
	SendAttachment(att tgbotapi.Chattable) (res tgbotapi.Message, err error)
}

type (
	Handler struct {
		cfg      *config.Config
		api      botApi
		services *Services
	}
	Services struct {
		Image        image.ImageService
		Subscription subscription.SubscriptionService
	}
)

func New(cfg *config.Config, botAPI botApi, services *Services) *Handler {
	h := &Handler{
		cfg:      cfg,
		api:      botAPI,
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

	res, err := h.api.SendAttachment(attachment)
	if err != nil {
		return
	}

	if file.TgID == "" {
		h.updateFile(ctx, file, res)
	}
}

func (h *Handler) CreateSubscription(ctx context.Context, message *tgbotapi.Message) error {
	inp, err := h.parseAndValidateSubscriptionInput(message)
	if err != nil {
		h.api.SendMessage(message.Chat.ID, err.Error())

		return err
	}

	err = h.services.Subscription.Create(ctx, inp, h.sendImage)
	if err != nil {
		log.Printf("Error creating subscription: %v", err)
		h.api.SendMessage(message.Chat.ID, "Error creating subscription!")

		return err
	}

	h.api.SendMessage(message.Chat.ID, "Subscription created successfully!")

	return nil
}

func (h *Handler) GetSubscription(ctx context.Context, message *tgbotapi.Message) {
	sub, err := h.services.Subscription.Get(ctx, message.Chat.ID)
	if err != nil {
		msgText := "Error getting subscription :d"

		var notFoundErr *custom_errors.NotFoundError
		if errors.As(err, &notFoundErr) {
			msgText = "No active subscription found!"
		}

		h.api.SendMessage(message.Chat.ID, msgText)

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

	h.api.SendMessage(message.Chat.ID, msgText)
}

func (h *Handler) DeleteSubscription(ctx context.Context, message *tgbotapi.Message) {
	sub, err := h.services.Subscription.Get(ctx, message.Chat.ID)
	if err != nil {
		msgText := "Error getting subscription :d"

		var notFoundErr *custom_errors.NotFoundError
		if errors.As(err, &notFoundErr) {
			msgText = "No active subscription found!"
		}

		h.api.SendMessage(message.Chat.ID, msgText)

		return
	}

	err = h.services.Subscription.Delete(ctx, sub.ChatId)
	if err != nil {
		h.api.SendMessage(message.Chat.ID, "Can not delete subscription :d")

		return
	}

	h.api.SendMessage(message.Chat.ID, "Subscription deleted successfully!")
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

	res, err := h.api.SendAttachment(attachment)
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
	rawMsg := strings.ReplaceAll(message.Text, " ", "")
	rawMsg = strings.ToLower(rawMsg)

	period, err := time.ParseDuration(rawMsg)
	if err != nil {
		errText := "Please enter a period in format like 1h30m\n" +
			fmt.Sprintf(
				"Hint: minimum: %s, maximun: %s",
				time_string.ShortDur(h.cfg.MinSubscriptionInterval),
				time_string.ShortDur(h.cfg.MaxSubscriptionInterval),
			)
		err = errors.New(errText)

		return domain.Subscription{}, err
	}

	if period.Seconds() < h.cfg.MinSubscriptionInterval.Seconds() ||
		period.Seconds() > h.cfg.MaxSubscriptionInterval.Seconds() {
		errText := fmt.Sprintf(
			"Subscription period must be between %s and %s!",
			time_string.ShortDur(h.cfg.MinSubscriptionInterval),
			time_string.ShortDur(h.cfg.MaxSubscriptionInterval),
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
