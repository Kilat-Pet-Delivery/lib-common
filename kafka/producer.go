package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer wraps kafka-go writer for publishing messages.
//
// Publish is called from HTTP handlers, so writers is guarded: a concurrent map
// write is a fatal runtime error in Go, not a recoverable race. See KPD-66.
type Producer struct {
	mu      sync.RWMutex
	writers map[string]*kafka.Writer
	brokers []string
	logger  *zap.Logger
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers []string, logger *zap.Logger) *Producer {
	return &Producer{
		writers: make(map[string]*kafka.Writer),
		brokers: brokers,
		logger:  logger,
	}
}

// getWriter returns or creates a writer for the given topic. Safe for concurrent
// use: the common case takes a read lock, and the miss path double-checks under
// the write lock so two goroutines racing on the same topic still share one
// writer rather than leaking a second one.
func (p *Producer) getWriter(topic string) *kafka.Writer {
	p.mu.RLock()
	w, exists := p.writers[topic]
	p.mu.RUnlock()
	if exists {
		return w
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if w, exists := p.writers[topic]; exists {
		return w
	}
	w = &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,

		// Without this, kafka-go refuses to publish to a topic that does not
		// exist yet and fails with "Unknown Topic Or Partition" -- even when the
		// broker has auto-creation enabled, because topic creation is driven by
		// the client. Consumers create topics on subscribe, producers do not,
		// which is why only the topics someone happened to consume ever existed.
		// See KPD-64.
		AllowAutoTopicCreation: true,
	}
	p.writers[topic] = w
	return w
}

// Publish sends a message to a Kafka topic.
func (p *Producer) Publish(ctx context.Context, topic, key string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	writer := p.getWriter(topic)
	msg := kafka.Message{
		Key:   []byte(key),
		Value: data,
		Time:  time.Now().UTC(),
	}

	if err := writer.WriteMessages(ctx, msg); err != nil {
		p.logger.Error("failed to publish message",
			zap.String("topic", topic),
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish to %s: %w", topic, err)
	}

	p.logger.Debug("message published",
		zap.String("topic", topic),
		zap.String("key", key),
	)
	return nil
}

// PublishEvent publishes a CloudEvent to a topic.
func (p *Producer) PublishEvent(ctx context.Context, topic string, event CloudEvent) error {
	return p.Publish(ctx, topic, event.ID, event)
}

// Close closes all writers.
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for topic, w := range p.writers {
		if err := w.Close(); err != nil {
			p.logger.Error("failed to close writer", zap.String("topic", topic), zap.Error(err))
		}
	}
	return nil
}
