package consumers

import (
	"encoding/json"

	"github.com/Dailiduzhou/library_manage_sys/internal/events"
	"github.com/Dailiduzhou/library_manage_sys/pkg/logger"
)

// BorrowEventHandler consumes borrow/return events.
type BorrowEventHandler struct{}

func NewBorrowEventHandler() *BorrowEventHandler {
	return &BorrowEventHandler{}
}

func (h *BorrowEventHandler) HandleMessage(topic string, partition int32, offset int64, value []byte) error {
	var evt events.BorrowEvent
	if err := json.Unmarshal(value, &evt); err != nil {
		logger.Infof("kafka: failed to unmarshal borrow event: %v (topic=%s partition=%d offset=%d)", err, topic, partition, offset)
		return err
	}

	logger.Infof("kafka: borrow event consumed: topic=%s partition=%d offset=%d event=%+v", topic, partition, offset, evt)
	return nil
}
