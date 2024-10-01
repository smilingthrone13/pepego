package domain

import "time"

type Subscription struct {
	ChatId    int64
	CreatedAt int64
	Period    int
}

func (s Subscription) GetSubscribedAtAsUnixTime() time.Time {
	return time.Unix(s.CreatedAt, 0)
}

func (s Subscription) GetPeriodAsDuration() time.Duration {
	return time.Duration(s.Period) * time.Second // todo: switch to hours
}
