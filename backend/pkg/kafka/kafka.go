package kafka

import (
	"fmt"
	"strings"

	"github.com/Dailiduzhou/library_manage_sys/internal/ports"
	"github.com/IBM/sarama"
)

var (
	defaultProducer ports.Producer
	consumerGroup   sarama.ConsumerGroup
)

// NewConfig creates a Sarama config with the given Kafka version.
func NewConfig(version string) (*sarama.Config, error) {
	cfg := sarama.NewConfig()
	parsedVersion, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		return nil, fmt.Errorf("invalid kafka version %q: %w", version, err)
	}
	cfg.Version = parsedVersion
	cfg.Producer.Return.Successes = true
	return cfg, nil
}

// InitProducer initializes a global SyncProducer and returns a Producer wrapper.
func InitProducer(brokers []string, cfg *sarama.Config) (ports.Producer, error) {
	producer, err := sarama.NewSyncProducer(normalizeBrokers(brokers), cfg)
	if err != nil {
		return nil, err
	}
	defaultProducer = &saramaProducer{producer: producer}
	return defaultProducer, nil
}

// InitConsumerGroup initializes a global ConsumerGroup.
func InitConsumerGroup(brokers []string, groupID string, cfg *sarama.Config) (sarama.ConsumerGroup, error) {
	group, err := sarama.NewConsumerGroup(normalizeBrokers(brokers), groupID, cfg)
	if err != nil {
		return nil, err
	}
	consumerGroup = group
	return consumerGroup, nil
}

// Close closes initialized Kafka clients.
func Close() error {
	var closeErr error
	if defaultProducer != nil {
		if p, ok := defaultProducer.(*saramaProducer); ok && p.producer != nil {
			if err := p.producer.Close(); err != nil {
				closeErr = err
			}
		}
	}
	if consumerGroup != nil {
		if err := consumerGroup.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	return closeErr
}

type saramaProducer struct {
	producer sarama.SyncProducer
}

func (p *saramaProducer) SendMessage(topic, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := p.producer.SendMessage(msg)
	return err
}

func normalizeBrokers(brokers []string) []string {
	result := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		trimmed := strings.TrimSpace(broker)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
