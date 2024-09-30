package subscription

import (
	"apubot/internal/config"
	"apubot/internal/domain"
	"apubot/pkg/custom_errors"
	"context"
	"github.com/pkg/errors"
	"log"
	"sync"
	"time"
)

type Service struct {
	cfg                  *config.Config
	repo                 SubscriptionRepository
	runningSubscriptions map[string]chan struct{} // todo: change to struct with closeChan and lastSentPicID?
	mu                   sync.RWMutex
}

func New(cfg *config.Config, repo SubscriptionRepository) *Service {
	service := &Service{
		cfg:                  cfg,
		repo:                 repo,
		runningSubscriptions: make(map[string]chan struct{}),
		mu:                   sync.RWMutex{},
	}

	err := service.scheduleExisting(context.Background())
	if err != nil {
		log.Fatalf("can not initialize Subscription service: %v", err)
	}

	return service
}

func (s *Service) getAllFromDB(ctx context.Context) (subs []domain.Subscription, err error) {
	subs, err = s.repo.GetAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "can not get all subscriptions")
	}

	return subs, nil
}

func (s *Service) scheduleExisting(ctx context.Context) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existingSubs, err := s.getAllFromDB(ctx)
	if err != nil {
		return errors.Wrap(err, "can not get existing subscriptions")
	}

	for i := range existingSubs {
		exitChan := make(chan struct{}, 1)

		// todo: start goroutine

		s.runningSubscriptions[existingSubs[i].ChatId] = exitChan
	}

	return nil
}

func (s *Service) startWorker(sub domain.Subscription) {
	subscribedAd := time.Unix(sub.CreatedAt, 0)
	timeSinceSubscribe := time.Since(subscribedAd)
	period := time.Duration(sub.Period) * time.Hour
	passedIntervals := timeSinceSubscribe / period

	nextRun := subscribedAd.Add((passedIntervals + 1) * period)
	delay := time.Until(nextRun)

	time.Sleep(delay)

	// todo: start goroutine
}

func (s *Service) Get(ctx context.Context, chatId string) (sub domain.Subscription, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.runningSubscriptions[chatId]
	if !ok {
		return sub, custom_errors.NewNotFound("can not find subscription")
	}

	sub, err = s.repo.Get(ctx, chatId)
	if err != nil {
		return sub, errors.Wrap(err, "can not get subscription")
	}

	return sub, nil
}

func (s *Service) Create(ctx context.Context, sub domain.Subscription) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// kill existing subscription goroutine if exists
	exitChanOld, ok := s.runningSubscriptions[sub.ChatId]
	if ok {
		exitChanOld <- struct{}{}
	}

	err := s.repo.Create(ctx, sub)
	if err != nil {
		return errors.Wrap(err, "can not create subscription")
	}

	exitChan := make(chan struct{})

	// todo: start goroutine

	s.runningSubscriptions[sub.ChatId] = exitChan

	return nil
}

func (s *Service) Delete(ctx context.Context, chatId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	exitChan, ok := s.runningSubscriptions[chatId]
	if !ok {
		return custom_errors.NewNotFound("can not find subscription")
	}

	err := s.repo.Delete(ctx, chatId)
	if err != nil {
		return errors.Wrap(err, "can not delete subscription")
	}

	exitChan <- struct{}{}

	delete(s.runningSubscriptions, chatId)

	return nil
}
