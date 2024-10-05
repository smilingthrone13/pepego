package subscription

import (
	"apubot/internal/config"
	"apubot/internal/domain"
	"apubot/pkg/custom_errors"
	"apubot/pkg/utils/queue"
	"context"
	"github.com/pkg/errors"
	"log"
	"sync"
	"time"
)

type (
	Service struct {
		cfg                  *config.Config
		repo                 SubscriptionRepository
		runningSubscriptions map[int64]chan struct{}
		mu                   sync.RWMutex
	}
)

func New(cfg *config.Config, repo SubscriptionRepository) *Service {
	service := &Service{
		cfg:                  cfg,
		repo:                 repo,
		runningSubscriptions: make(map[int64]chan struct{}),
		mu:                   sync.RWMutex{},
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

func (s *Service) startWorker(
	sub domain.Subscription,
	exitChan chan struct{},
	sendFunc func(chatId int64, q *queue.Queue) error,
) {
	passedIntervals := time.Since(sub.SubscribedAtAsUnixTime()) / sub.PeriodAsDurationInSeconds()
	nextRun := sub.SubscribedAtAsUnixTime().Add((passedIntervals + 1) * sub.PeriodAsDurationInSeconds())

	workerInput := &StartWorkerInput{
		ChatID:   sub.ChatId,
		ExitChan: exitChan,
		Delay:    time.Until(nextRun),
		Period:   sub.PeriodAsDurationInSeconds(),
	}

	go s.startSubscription(workerInput, sendFunc)
}

func (s *Service) startSubscription(
	inp *StartWorkerInput,
	sendFunc func(chatId int64, q *queue.Queue) error,
) {
	failCount := 0
	timeout := inp.Delay // initial delay before next scheduled event
	q := queue.NewQueue(s.cfg.LastSentQueueSize)

	for {
		select {
		case <-time.After(timeout):
		case <-inp.ExitChan:
			return
		}

		start := time.Now()

		if failCount >= s.cfg.MaxRetries {
			log.Printf("Max retries reached for chat %d, auto-deleting subscription!", inp.ChatID)
			err := s.Delete(context.Background(), inp.ChatID)
			if err != nil {
				log.Printf("Can not auto-delete subscription %d: %v", inp.ChatID, err)
			}

			return
		}

		err := sendFunc(inp.ChatID, q)
		timeout = inp.Period - time.Since(start) // schedule next event
		if err != nil {
			failCount++
			log.Printf(
				"Can not send scheduled message to chat %d (%d/%d): %v",
				inp.ChatID, failCount, s.cfg.MaxRetries, err,
			)

			continue
		}

		failCount = 0
	}
}

func (s *Service) RescheduleExisting(
	ctx context.Context,
	sendFunc func(chatId int64, q *queue.Queue) error,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existingSubs, err := s.getAllFromDB(ctx)
	if err != nil {
		return errors.Wrap(err, "can not reschedule existing subscriptions")
	}

	// kill all running subscriptions
	for _, ch := range s.runningSubscriptions {
		ch <- struct{}{}
		close(ch)
	}

	s.runningSubscriptions = make(map[int64]chan struct{}, len(existingSubs))

	for i := range existingSubs {
		exitChan := make(chan struct{}, 1)

		go s.startWorker(existingSubs[i], exitChan, sendFunc)

		s.runningSubscriptions[existingSubs[i].ChatId] = exitChan
	}

	log.Printf("Rescheduled %d existing subscription(s)!", len(s.runningSubscriptions))

	return nil
}

func (s *Service) Get(ctx context.Context, chatId int64) (sub domain.Subscription, err error) {
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

func (s *Service) Create(
	ctx context.Context,
	sub domain.Subscription,
	sendFunc func(chatId int64, q *queue.Queue) error,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// kill running subscription goroutine if exists
	exitChanOld, ok := s.runningSubscriptions[sub.ChatId]
	if ok {
		exitChanOld <- struct{}{}
		close(exitChanOld)
	}

	err := s.repo.Create(ctx, sub)
	if err != nil {
		return errors.Wrap(err, "can not create subscription")
	}

	exitChan := make(chan struct{}, 1)

	workerInput := &StartWorkerInput{
		ChatID:   sub.ChatId,
		ExitChan: exitChan,
		Delay:    time.Duration(1) * time.Second,
		Period:   sub.PeriodAsDurationInSeconds(),
	}

	go s.startSubscription(workerInput, sendFunc)

	s.runningSubscriptions[sub.ChatId] = exitChan

	return nil
}

func (s *Service) Delete(ctx context.Context, chatId int64) error {
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

	exitChan <- struct{}{} // stop running subscription goroutine
	close(exitChan)

	delete(s.runningSubscriptions, chatId)

	return nil
}
