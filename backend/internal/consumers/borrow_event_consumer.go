package consumers

import (
	"encoding/json"
	"log"

	"github.com/Dailiduzhou/library_manage_sys/internal/events"
	"github.com/IBM/sarama"
)

// BorrowEventHandler consumes borrow/return events.
type BorrowEventHandler struct{}

func NewBorrowEventHandler() *BorrowEventHandler {
	return &BorrowEventHandler{}
}

func (h *BorrowEventHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *BorrowEventHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *BorrowEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var evt events.BorrowEvent
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			log.Printf("kafka: failed to unmarshal borrow event: %v (topic=%s partition=%d offset=%d)", err, msg.Topic, msg.Partition, msg.Offset)
			session.MarkMessage(msg, "")
			continue
		}

		log.Printf("kafka: borrow event consumed: topic=%s partition=%d offset=%d event=%+v", msg.Topic, msg.Partition, msg.Offset, evt)
		session.MarkMessage(msg, "")
	}
	return nil
}
