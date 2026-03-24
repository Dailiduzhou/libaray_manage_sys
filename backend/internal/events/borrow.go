package events

import "time"

const (
	BorrowedEventType = "borrowed"
	ReturnedEventType = "returned"
)

// BorrowEvent represents a borrow/return domain event.
type BorrowEvent struct {
	EventType  string    `json:"event_type"`
	BorrowID   uint      `json:"borrow_id"`
	UserID     uint      `json:"user_id"`
	BookID     uint      `json:"book_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
