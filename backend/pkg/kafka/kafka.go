package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/Dailiduzhou/library_manage_sys/internal/ports"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	defaultProducer ports.Producer
	producerClient  *kgo.Client
	consumerClient  *kgo.Client
)

// Config keeps Kafka settings for compatibility during client migration.
type Config struct {
	Version string
}

// NewConfig keeps the existing version input path for compatibility.
func NewConfig(version string) (*Config, error) {
	if strings.TrimSpace(version) == "" {
		return nil, fmt.Errorf("kafka version cannot be empty")
	}
	return &Config{Version: version}, nil
}

// InitProducer initializes a global Kafka producer and returns a Producer wrapper.
func InitProducer(brokers []string, _ *Config) (ports.Producer, error) {
	normalized := normalizeBrokers(brokers)
	client, err := kgo.NewClient(
		kgo.SeedBrokers(normalized...),
	)
	if err != nil {
		return nil, err
	}

	producerClient = client
	defaultProducer = &franzProducer{client: producerClient}
	return defaultProducer, nil
}

// InitConsumerGroup initializes a global consumer client.
func InitConsumerGroup(brokers []string, groupID string, topics []string, _ *Config) (*kgo.Client, error) {
	normalizedBrokers := normalizeBrokers(brokers)
	normalizedTopics := normalizeBrokers(topics)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(normalizedBrokers...),
		kgo.ConsumerGroup(strings.TrimSpace(groupID)),
		kgo.ConsumeTopics(normalizedTopics...),
		kgo.DisableAutoCommit(),
	)
	if err != nil {
		return nil, err
	}

	consumerClient = client
	return consumerClient, nil
}

// Close closes initialized Kafka clients.
func Close() error {
	var closeErr error
	if producerClient != nil {
		producerClient.Close()
		producerClient = nil
	}
	if consumerClient != nil {
		consumerClient.Close()
		consumerClient = nil
	}
	if defaultProducer != nil {
		if p, ok := defaultProducer.(*franzProducer); ok && p.client != nil {
			p.client = nil
		}
		defaultProducer = nil
	}

	return closeErr
}

type franzProducer struct {
	client *kgo.Client
}

func (p *franzProducer) SendMessage(topic, key string, value []byte) error {
	if p.client == nil {
		return fmt.Errorf("kafka producer is not initialized")
	}

	msg := &kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}

	return p.client.ProduceSync(context.Background(), msg).FirstErr()
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
