package domain

import "context"

const (
	EventDepositCreated    = "deposit.created"
	EventWithdrawalCreated = "withdrawal.created"
)

type Event struct {
	Topic string
	Key   string
	Value any
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}
