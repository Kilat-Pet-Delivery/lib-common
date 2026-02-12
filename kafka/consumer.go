package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// MessageHandler processes a consumed Kafka message.
type MessageHandler func(ctx context.Context, msg kafka.Message) error

// Consumer wraps kafka-go reader for consuming messages.
type Consumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	topic   string
	groupID string
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(brokers []string, groupID, topic string, logger *zap.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 1,
		MaxBytes: 10e6, // 10MB
	})

	return &Consumer{
		reader:  reader,
		logger:  logger,
		topic:   topic,
		groupID: groupID,
	}
}

// Consume starts consuming messages and delegates to the handler.
// It blocks until the context is cancelled.
func (c *Consumer) Consume(ctx context.Context, handler MessageHandler) error {
	c.logger.Info("starting consumer",
		zap.String("topic", c.topic),
		zap.String("group", c.groupID),
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer stopping", zap.String("topic", c.topic))
			return ctx.Err()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				c.logger.Error("failed to fetch message",
					zap.String("topic", c.topic),
					zap.Error(err),
				)
				continue
			}

			if err := handler(ctx, msg); err != nil {
				c.logger.Error("failed to handle message",
					zap.String("topic", c.topic),
					zap.Int64("offset", msg.Offset),
					zap.Error(err),
				)
				continue
			}

			// Commit only on successful processing
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("failed to commit message",
					zap.String("topic", c.topic),
					zap.Int64("offset", msg.Offset),
					zap.Error(err),
				)
			}
		}
	}
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close consumer for topic %s: %w", c.topic, err)
	}
	return nil
}
