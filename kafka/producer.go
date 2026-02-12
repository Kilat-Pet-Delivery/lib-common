package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer wraps kafka-go writer for publishing messages.
type Producer struct {
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

// getWriter returns or creates a writer for the given topic.
func (p *Producer) getWriter(topic string) *kafka.Writer {
	if w, exists := p.writers[topic]; exists {
		return w
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
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
	for topic, w := range p.writers {
		if err := w.Close(); err != nil {
			p.logger.Error("failed to close writer", zap.String("topic", topic), zap.Error(err))
		}
	}
	return nil
}
