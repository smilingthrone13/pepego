package domain

type Subscription struct {
	ChatId    string
	CreatedAt int64
	Period    int
}

type CreateSubscriptionDTO struct {
	ChatId string
	Period int
}
