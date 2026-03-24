package ports

// Producer is a minimal event producer abstraction for services.
type Producer interface {
	SendMessage(topic, key string, value []byte) error
}
