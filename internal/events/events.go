package events

import "context"

type ExpenseCreatedEvent struct {
	EventType string  `json:"event_type"`
	ExpenseId uint    `json:"expense_id"`
	UserID    uint    `json:"user_id"`
	TravelID  uint    `json:"travel_id"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	CreatedAt string  `json:"created_at"`
}

type Publisher interface {
	PublishExpense(ctx context.Context, event ExpenseCreatedEvent) error
}

type NoopPublisher struct{}

func (n *NoopPublisher) PublishExpense(ctx context.Context, event ExpenseCreatedEvent) error {
	return nil
}
