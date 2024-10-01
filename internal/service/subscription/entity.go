package subscription

import "time"

type StartWorkerInput struct {
	ChatID   int64
	ExitChan chan struct{}
	Delay    time.Duration
	Period   time.Duration
}

// todo: add last used pic queue
