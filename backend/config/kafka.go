package config

import (
	"strings"
)

// KafkaConfig holds Kafka-related configuration.
type KafkaConfig struct {
	Brokers       []string
	Version       string
	GroupID       string
	BorrowedTopic string
	ReturnedTopic string
}

// LoadKafkaConfig loads Kafka config from environment variables.
func LoadKafkaConfig() KafkaConfig {
	brokersRaw := getEnv("KAFKA_BROKERS", "localhost:9092")
	brokers := strings.Split(brokersRaw, ",")

	return KafkaConfig{
		Brokers:       brokers,
		Version:       getEnv("KAFKA_VERSION", "2.8.1"),
		GroupID:       getEnv("KAFKA_GROUP_ID", "library-borrow-consumer"),
		BorrowedTopic: getEnv("KAFKA_TOPIC_BORROWED", "borrowed-events"),
		ReturnedTopic: getEnv("KAFKA_TOPIC_RETURNED", "returned-events"),
	}
}
